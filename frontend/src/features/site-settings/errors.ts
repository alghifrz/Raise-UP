import { isApiError } from '../../lib/api/types'

export function toSiteSettingsErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'SITE_SETTINGS_NOT_FOUND':
      return 'Pengaturan situs tidak ditemukan.'
    case 'INVALID_REQUEST':
      return error.message || 'Data yang dikirim tidak valid.'
    case 'NETWORK_ERROR':
      return 'Tidak dapat terhubung ke server. Periksa koneksi Anda.'
    case 'UNAUTHORIZED':
    case 'INVALID_CREDENTIALS':
      return 'Sesi Anda telah berakhir. Silakan masuk kembali.'
    case 'FORBIDDEN':
      return 'Anda tidak memiliki izin untuk melakukan tindakan ini.'
    default:
      if (error.status === 400) {
        return error.message || 'Data yang dikirim tidak valid.'
      }
      return 'Gagal menyimpan pengaturan situs.'
  }
}
