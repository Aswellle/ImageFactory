// Shared password utilities used across auth views.

/**
 * Score password strength on 5 criteria:
 * - At least 8 characters
 * - Contains lowercase letter
 * - Contains uppercase letter
 * - Contains number
 * - Contains special symbol
 *
 * Returns 0-5. A score of 3 (Medium) is the minimum for registration.
 */
export function passwordScore(pw: string): number {
  let score = 0
  if (pw.length >= 8) score++
  if (/[a-z]/.test(pw)) score++
  if (/[A-Z]/.test(pw)) score++
  if (/\d/.test(pw)) score++
  if (/[^a-zA-Z0-9]/.test(pw)) score++
  return score
}

/** Minimum score required for registration (Medium strength). */
export const MIN_PASSWORD_SCORE = 3

/** Email placeholder shown in auth forms (avoids vue-i18n @ parsing issue). */
export const EMAIL_PLACEHOLDER = 'user@example.com'
