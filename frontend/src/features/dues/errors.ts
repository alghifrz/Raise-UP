import { isApiError } from '../../lib/api/types'

export function toDuesErrorMessage(error: unknown): string {
  if (!isApiError(error)) {
    return 'Terjadi kesalahan. Silakan coba lagi.'
  }

  switch (error.code) {
    case 'DUES_PERIOD_NOT_FOUND':
      return 'Periode iuran tidak ditemukan.'
    case 'DUES_PAYMENT_NOT_FOUND':
      return 'Pembayaran iuran tidak ditemukan.'
    case 'RESIDENT_NOT_FOUND':
      return 'Warga tidak ditemukan.'
    case 'DUES_PERIOD_ALREADY_EXISTS':
      return 'Periode iuran tersebut sudah ada.'
    case 'DUES_PAYMENT_ALREADY_EXISTS':
      return 'Pembayaran untuk warga dan periode ini sudah ada.'
    case 'INVALID_PAYMENT_AMOUNT':
      return 'Nominal pembayaran harus sama dengan nominal periode.'
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
        return 'Data iuran tidak ditemukan.'
      }
      if (error.status === 409) {
        return error.message || 'Operasi tidak dapat dilanjutkan karena konflik data.'
      }
      return 'Terjadi kesalahan. Silakan coba lagi.'
  }
}
