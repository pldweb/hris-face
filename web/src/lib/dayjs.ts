// Single source of truth for date/time in this app: everything is interpreted
// in the business's own timezone, Asia/Jakarta -- matching api/internal/config
// (TIMEZONE env, default Asia/Jakarta) and the Postgres session timezone pinned
// in api/internal/db.Connect.
//
// Plain `dayjs(iso).format()` uses the VIEWER's own browser/OS timezone, not
// the business's -- and dayjs.tz.setDefault() does NOT change that for bare
// dayjs() calls, only for explicit .tz() ones (verified empirically: a browser
// set to America/New_York still showed dayjs().format() in New York time even
// after setDefault('Asia/Jakarta') was called; only dayjs().tz() picked it up).
// So every date/time in this app goes through this wrapper instead of
// importing 'dayjs' directly, rather than trusting a plugin default that turned
// out not to apply where we needed it.
import dayjs from 'dayjs'
import 'dayjs/locale/id'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'

dayjs.extend(utc)
dayjs.extend(timezone)
dayjs.locale('id')

export const JAKARTA_TZ = 'Asia/Jakarta'

type DayjsArgs = Parameters<typeof dayjs>

/** Drop-in dayjs() that always resolves in Asia/Jakarta, whatever the viewer's
 *  own device is set to. Use this everywhere instead of the raw package. */
export default function tzDayjs(...args: DayjsArgs) {
  return dayjs(...args).tz(JAKARTA_TZ)
}

export type Dayjs = ReturnType<typeof tzDayjs>
