// Validates that a string only contains letters, numbers, underscores, and dots
// Useful for ensuring usernames or slugs are safe
export const isValidUsername = (username: string): boolean => {
    const regex = /^[a-zA-Z0-9_.]+$/;
    return regex.test(username);
};

// Sanitizes input to remove potential script tags (basic)
// Note: React handles most XSS protection, but this can be useful for other contexts
export const sanitizeInput = (input: string): string => {
    return input.replace(/<[^>]*>?/gm, '');
};

// Validates a basic email format
export const isValidEmail = (email: string): boolean => {
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return regex.test(email);
};
