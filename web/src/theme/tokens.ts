// Single source of truth for Ant Design Vue tokens.
// docs/PRD.md section 10.2: Restrained color strategy, institutional not decorative.
// Overriding colorPrimary is the one change that most determines identity --
// the AntD default (#1677FF) reads as an AntD demo, not this product.
import type { ThemeConfig } from 'ant-design-vue/es/config-provider/context'

export const themeTokens: ThemeConfig = {
  token: {
    colorPrimary: '#21409A',
    colorInfo: '#21409A',
    colorBgLayout: '#F5F6F8',
    colorBgContainer: '#FFFFFF',
    colorText: '#172554',
    colorTextSecondary: '#64748B',
    colorBorder: '#E2E8F0',
    borderRadius: 8,
    controlHeight: 40,
    controlHeightLG: 48,
    controlHeightSM: 32,
    fontSize: 14,
    boxShadowSecondary: '0 16px 40px rgba(23, 37, 84, 0.08)',
    // "Google Sans" itself is Google's internal typeface (not on Google Fonts,
    // not publicly licensed); Plus Jakarta Sans is the closest free substitute --
    // see web/index.html for the stylesheet link. System stack stays as fallback.
    fontFamily:
      "'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  },
}

// Attendance status colors are locked to AntD semantic defaults and used
// ONLY for kehadiran meaning (on_time/late/absent), never as decoration
// elsewhere on the surface -- this is what lets HR scan the table by color alone.
export const attendanceStatusColor: Record<string, string> = {
  on_time: 'success',
  late: 'warning',
  early_leave: 'warning',
  absent: 'error',
}
