// Validates that a string only contains letters, numbers, underscores, and dots
// Useful for ensuring usernames or slugs are safe
export const isValidUsername = (username: string): boolean => {
    const regex = /^[a-zA-Z0-9_.]+$/;
    return regex.test(username);
};

// Checks if a string contains any HTML-like tags (< or >)
export const containsHtml = (input: string): boolean => {
    return /[<>]/.test(input);
};

// Checks for characters that could be problematic (e.g., % for SQL LIKE, \ for escaping)
export const containsForbiddenChars = (input: string): boolean => {
    return /[%\\]/.test(input);
};

// Sanitizes input to remove all HTML tags and specific problematic characters like '%'
// This is used for general text fields to prevent basic XSS and SQL LIKE issues
export const sanitizeInput = (input: string): string => {
    return input
        .replace(/<[^>]*>?/gm, '') // Remove all HTML tags
        .replace(/%/g, '');        // Remove % character
};

// Validates a basic email format
export const isValidEmail = (email: string): boolean => {
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return regex.test(email);
};
