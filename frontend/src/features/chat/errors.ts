import { isApiError } from '../../lib/api/types'

export function toChatErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'CONVERSATION_NOT_FOUND':
      return 'Percakapan tidak ditemukan.'
    case 'USER_NOT_FOUND':
      return 'Pengguna yang dipilih tidak ditemukan.'
    case 'FORBIDDEN':
      return 'Anda bukan peserta percakapan ini.'
    case 'WHATSAPP_DISABLED':
      return 'WhatsApp belum dikonfigurasi di server.'
    case 'WHATSAPP_SEND_FAILED':
      return error.message || 'Gagal mengirim pesan WhatsApp.'
    case 'INVALID_REQUEST':
      return error.message || 'Data yang dikirim tidak valid.'
    case 'NETWORK_ERROR':
      return 'Tidak dapat terhubung ke server. Periksa koneksi Anda.'
    case 'UNAUTHORIZED':
    case 'INVALID_CREDENTIALS':
      return 'Sesi Anda telah berakhir. Silakan masuk kembali.'
    default:
      if (error.status === 400) {
        return error.message || 'Data yang dikirim tidak valid.'
      }
      if (error.status === 403) {
        return 'Anda bukan peserta percakapan ini.'
      }
      if (error.status === 404) {
        return 'Data chat tidak ditemukan.'
      }
      if (error.status === 502 || error.status === 503) {
        return error.message || 'Layanan WhatsApp tidak tersedia.'
      }
      return 'Terjadi kesalahan. Silakan coba lagi.'
  }
}
