import { isApiError } from '../../lib/api/types'

export function toAnnouncementErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'ANNOUNCEMENT_NOT_FOUND':
      return 'Pengumuman tidak ditemukan.'
    case 'RESIDENT_NOT_FOUND':
      return 'Salah satu penerima warga tidak ditemukan.'
    case 'ANNOUNCEMENT_RECIPIENT_REQUIRED':
      return 'Pengumuman privat harus memiliki minimal satu penerima.'
    case 'INVALID_STATUS_TRANSITION':
      return 'Status pengumuman sudah tidak dapat diubah.'
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
      if (error.status === 404) {
        return 'Pengumuman tidak ditemukan.'
      }
      if (error.status === 409) {
        return 'Operasi tidak dapat dilanjutkan karena konflik data.'
      }
      return 'Terjadi kesalahan. Silakan coba lagi.'
  }
}
