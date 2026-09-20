import { isApiError } from '../../lib/api/types'

export function toGalleryErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'GALLERY_ITEM_NOT_FOUND':
      return 'Foto tidak ditemukan.'
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
        return 'Foto tidak ditemukan.'
      }
      return 'Terjadi kesalahan. Silakan coba lagi.'
  }
}
