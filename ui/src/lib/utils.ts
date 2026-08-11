import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Snippet } from 'svelte';
import type { HTMLAttributes, HTMLInputAttributes, HTMLTextareaAttributes } from 'svelte/elements';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export type WithElementRef<
	T,
	Ref extends string = 'ref',
	RefElement = T extends { ref?: infer R } ? R : HTMLElement | null
> = Omit<T, Ref> & {
	[K in Ref]?: RefElement;
};

export type WithoutChild<T> = T extends { children: infer _ } ? Omit<T, 'children'> : T;

export type WithChild<
	T,
	U extends string = 'children',
	V = T extends { [K in U]: infer _ } ? T[U] : never
> = WithoutChild<T> & {
	[K in U]: V;
};
