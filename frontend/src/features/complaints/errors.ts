import { isApiError } from '../../lib/api/types'

export function toComplaintErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'COMPLAINT_NOT_FOUND':
      return 'Pengaduan tidak ditemukan.'
    case 'RESIDENT_NOT_FOUND':
      return 'Warga yang dipilih tidak ditemukan.'
    case 'INVALID_STATUS_TRANSITION':
      return 'Perubahan status tidak diizinkan untuk pengaduan ini.'
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
      if (error.status === 404) {
        return 'Pengaduan tidak ditemukan.'
      }
      if (error.status === 409) {
        return 'Operasi tidak dapat dilanjutkan karena konflik data.'
      }
      return 'Terjadi kesalahan. Silakan coba lagi.'
  }
}
