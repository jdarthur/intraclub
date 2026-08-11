// Photo API client. Photo records are backed by the existing REST CRUD routes
// generated from `model.Photo` (see main.go):
//
//	GET    /api/photo     -> { resource: Photo[] }
//	GET    /api/photo/:id -> { resource: Photo }
//	POST   /api/photo     -> { resource: Photo }  (owner = token user)
//	PUT    /api/photo/:id -> { resource: Photo }
//	DELETE /api/photo/:id -> { resource: Photo }
//
// A Photo is a binary image asset (e.g. a Facility layout photo) owned by a
// User. `contents` is base64-encoded image bytes (Go's `[]byte` marshals to
// base64 over JSON); `file_type` is the PhotoType enum value. Record IDs are
// 16-char hex strings; an empty string represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

// PhotoType enum values mirror model.PhotoType (model/photo.go).
export const PHOTO_TYPE_PNG = 0;
export const PHOTO_TYPE_JPG = 1;
export const PHOTO_TYPE_JPEG = 2;
export const PHOTO_TYPE_GIF = 3;
export const PHOTO_TYPE_WEBP = 4;

export const photoTypeLabels: Record<number, string> = {
	[PHOTO_TYPE_PNG]: 'png',
	[PHOTO_TYPE_JPG]: 'jpg',
	[PHOTO_TYPE_JPEG]: 'jpeg',
	[PHOTO_TYPE_GIF]: 'gif',
	[PHOTO_TYPE_WEBP]: 'webp'
};

export type PhotoType = number;

export interface Photo {
	id: string;
	owner: string;
	alt_text: string;
	contents: string; // base64-encoded image bytes
	file_type: PhotoType;
}

export interface PhotoInput {
	alt_text: string;
	contents: string; // base64-encoded image bytes
	file_type: PhotoType;
}

const BASE = '/api/photo';

// dataUrlFor returns a `data:` URL that can be passed straight to an <img src>,
// decoding the base64 contents back into a renderable image.
export function dataUrlFor(photo: Photo): string {
	const mime = photoTypeLabels[photo.file_type] ?? 'png';
	return `data:image/${mime};base64,${photo.contents}`;
}

// photoTypeFromExtension maps a file extension to its PhotoType value, or
// returns null when the extension isn't a supported image type.
export function photoTypeFromExtension(filename: string): PhotoType | null {
	const ext = filename.split('.').pop()?.toLowerCase() ?? '';
	switch (ext) {
		case 'png':
			return PHOTO_TYPE_PNG;
		case 'jpg':
			return PHOTO_TYPE_JPG;
		case 'jpeg':
			return PHOTO_TYPE_JPEG;
		case 'gif':
			return PHOTO_TYPE_GIF;
		case 'webp':
			return PHOTO_TYPE_WEBP;
		default:
			return null;
	}
}

// unwrap parses a CrudCommon response (`{ resource: ... }`) and throws the
// backend error message on non-2xx responses so callers can surface it.
async function unwrap(res: Response): Promise<any> {
	if (!res.ok) {
		let message = `Request failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// non-JSON error body; fall back to the status message
		}
		throw new Error(message);
	}
	const body = await res.json();
	return body?.resource;
}

export async function listPhotos(): Promise<Photo[]> {
	return unwrap(await authFetch(BASE));
}

export async function getPhoto(id: string): Promise<Photo> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export async function createPhoto(input: PhotoInput): Promise<Photo> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updatePhoto(id: string, input: PhotoInput): Promise<Photo> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deletePhoto(id: string): Promise<void> {
	const res = await authFetch(`${BASE}/${id}`, { method: 'DELETE' });
	if (!res.ok) {
		let message = `Delete failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// non-JSON error body; fall back to the status message
		}
		throw new Error(message);
	}
}
