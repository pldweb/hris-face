package employee

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ImportRow reports the outcome of one CSV line. Rows are reported individually
// rather than failing the whole file: HR pasting 80 employees should not lose
// 79 good rows because one has a duplicate NIK.
type ImportRow struct {
	Line         int    `json:"line"`
	NIK          string `json:"nik"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	TempPassword string `json:"temp_password,omitempty"`
	Error        string `json:"error,omitempty"`
}

type ImportResult struct {
	Created []ImportRow `json:"created"`
	Failed  []ImportRow `json:"failed"`
}

var ErrEmptyCSV = errors.New("file CSV kosong atau tidak punya baris data")

// ImportCSV expects a header row containing nik, nama, email and optionally
// departemen (matched by name, created if missing).
func (s *Service) ImportCSV(ctx context.Context, r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV tidak bisa dibaca: %w", err)
	}
	if len(records) < 2 {
		return nil, ErrEmptyCSV
	}

	idx, err := mapHeader(records[0])
	if err != nil {
		return nil, err
	}

	departments, err := s.departmentsByName(ctx)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{Created: []ImportRow{}, Failed: []ImportRow{}}

	for i, rec := range records[1:] {
		row := ImportRow{Line: i + 2}
		get := func(key string) string {
			pos, ok := idx[key]
			if !ok || pos >= len(rec) {
				return ""
			}
			return strings.TrimSpace(rec[pos])
		}

		row.NIK, row.FullName, row.Email = get("nik"), get("nama"), get("email")
		if row.NIK == "" || row.FullName == "" || row.Email == "" {
			row.Error = "nik, nama, dan email wajib diisi"
			result.Failed = append(result.Failed, row)
			continue
		}
		if !strings.Contains(row.Email, "@") {
			row.Error = "format email tidak valid"
			result.Failed = append(result.Failed, row)
			continue
		}

		var departmentID *string
		if name := get("departemen"); name != "" {
			id, ok := departments[strings.ToLower(name)]
			if !ok {
				created, err := s.CreateDepartment(ctx, name)
				if err != nil {
					row.Error = "gagal membuat departemen " + name
					result.Failed = append(result.Failed, row)
					continue
				}
				id = created.ID
				departments[strings.ToLower(name)] = id
			}
			departmentID = &id
		}

		out, err := s.Create(ctx, CreateEmployeeInput{
			NIK: row.NIK, FullName: row.FullName, Email: row.Email, DepartmentID: departmentID,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrEmailTaken):
				row.Error = "email sudah terdaftar"
			case errors.Is(err, ErrNIKTaken):
				row.Error = "NIK sudah terdaftar"
			default:
				row.Error = "gagal membuat karyawan"
			}
			result.Failed = append(result.Failed, row)
			continue
		}

		row.TempPassword = out.TempPassword
		result.Created = append(result.Created, row)
	}

	return result, nil
}

// mapHeader accepts a few common spellings so HR does not have to match an
// exact template to get their spreadsheet in.
func mapHeader(header []string) (map[string]int, error) {
	aliases := map[string]string{
		"nik": "nik", "nip": "nik", "id": "nik",
		"nama": "nama", "nama lengkap": "nama", "name": "nama", "full_name": "nama",
		"email": "email", "e-mail": "email", "surel": "email",
		"departemen": "departemen", "department": "departemen", "divisi": "departemen",
	}

	idx := map[string]int{}
	for i, raw := range header {
		key := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff")))
		if canonical, ok := aliases[key]; ok {
			idx[canonical] = i
		}
	}

	for _, required := range []string{"nik", "nama", "email"} {
		if _, ok := idx[required]; !ok {
			return nil, fmt.Errorf("kolom %q tidak ditemukan di header CSV", required)
		}
	}
	return idx, nil
}

func (s *Service) departmentsByName(ctx context.Context) (map[string]string, error) {
	list, err := s.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(list))
	for _, d := range list {
		out[strings.ToLower(d.Name)] = d.ID
	}
	return out, nil
}
