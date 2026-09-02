/**
 * Branded string types mirroring what the api actually serves on the wire.
 *
 * It's just to reduce cognitive load. Or atleast I think so
 */

export type UUID = string & { readonly brand: 'UUID' };
export type ISODateTimeString = string & { readonly brand: 'ISODateTimeString' };

export const asUuid = (value: string): UUID => value as UUID;
export const asIsoDateTime = (value: string): ISODateTimeString => value as ISODateTimeString;
