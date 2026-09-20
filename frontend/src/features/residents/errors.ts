import { isApiError } from '../../lib/api/types'

export function toResidentErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'RESIDENT_NOT_FOUND':
      return 'Data warga tidak ditemukan.'
    case 'RESIDENT_DELETE_CONFLICT':
      return 'Data warga tidak dapat dihapus karena masih digunakan oleh data lain.'
    case 'INVALID_REQUEST':
      return error.message || 'Data yang dikirim tidak valid.'
    case 'NETWORK_ERROR':
      return 'Tidak dapat terhubung ke server. Periksa koneksi Anda.'
    case 'UNAUTHORIZED':
    case 'INVALID_CREDENTIALS':
      return 'Sesi Anda telah berakhir. Silakan masuk kembali.'
    default:
      if (error.status === 400) {
        return 'Data yang dikirim tidak valid.'
      }
      return 'Terjadi kesalahan. Silakan coba lagi.'
  }
}
