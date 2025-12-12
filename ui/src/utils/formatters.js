export const formatPhoneNumber = (phoneNumber) => {
    if (!phoneNumber) return ''

    // Basic cleaning
    const cleaned = ('' + phoneNumber).replace(/\D/g, '')

    // US/CA Format (+1 or just 10 digits)
    // Check if it starts with 1 and is 11 digits
    if (cleaned.length === 11 && cleaned.startsWith('1')) {
        const match = cleaned.match(/^1(\d{3})(\d{3})(\d{4})$/)
        if (match) {
            return `+1 (${match[1]}) ${match[2]}-${match[3]}`
        }
    }

    // Check if it is 10 digits (assume US)
    if (cleaned.length === 10) {
        const match = cleaned.match(/^(\d{3})(\d{3})(\d{4})$/)
        if (match) {
            return `(${match[1]}) ${match[2]}-${match[3]}`
        }
    }

    // International or other fallback
    // If provided as +..., keep it, else add +
    if (phoneNumber.startsWith('+')) {
        return phoneNumber
    }

    return `+${cleaned}`
}

export const formatTime = (seconds) => {
    if (!seconds) return '0:00'
    const m = Math.floor(seconds / 60)
    const s = Math.floor(seconds % 60)
    return `${m}:${s < 10 ? '0' : ''}${s}`
}
