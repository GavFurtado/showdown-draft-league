import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

// Pure text-input utilities. Shared across features.
export function containsHtml(input: string): boolean {
  return /[<>]/.test(input);
}

export function containsForbiddenChars(input: string): boolean {
  return /[%\\]/.test(input);
}

export function sanitizeInput(input: string): string {
  return input.replace(/<[^>]*>?/gm, '').replace(/%/g, '');
}

export function htmlRejection(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null =>
    containsHtml(control.value) ? { containsHtml: true } : null;
}

export function forbiddenChars(): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null =>
    containsForbiddenChars(control.value) ? { forbiddenChars: true } : null;
}
