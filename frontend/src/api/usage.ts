import { api } from './client'
import type { UsageStat, UsageSummary } from '@/types'

// usageApi wraps the user usage dashboard endpoints. Both endpoints return the
// bare payload under the standard { data } envelope.
export const usageApi = {
  get(): Promise<UsageSummary> {
    return api.get('/usage').then((r) => r.data.data)
  },
  history(days = 30): Promise<UsageStat[]> {
    return api.get('/usage/history', { params: { days } }).then((r) => r.data.data)
  },
}
