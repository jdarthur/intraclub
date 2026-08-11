<script lang="ts">
	import { createPhoto, photoTypeFromExtension, photoTypeLabels } from '$lib/photo';
	import { goto } from '$app/navigation';
	import type { PhotoType } from '$lib/photo';

	let altText = $state('');
	let contents = $state('');
	let fileType = $state<PhotoType>(0);
	let error = $state('');
	let submitting = $state(false);

	function onFileChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		const type = photoTypeFromExtension(file.name);
		if (type === null) {
			error = `Unsupported file type. Use one of: ${Object.values(photoTypeLabels).join(', ')}.`;
			return;
		}
		fileType = type;
		error = '';
		const reader = new FileReader();
		reader.onload = () => {
			const result = reader.result as string;
			contents = result.split(',')[1] ?? '';
		};
		reader.readAsDataURL(file);
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		if (!contents) {
			error = 'Please choose an image file to upload.';
			return;
		}
		submitting = true;
		try {
			const created = await createPhoto({ alt_text: altText, contents, file_type: fileType });
			await goto(`/photos/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create photo';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New photo</title>
</svelte:head>

<h1>New photo</h1>
<a href="/photos">&larr; Back to photos</a>

<form onsubmit={handleSubmit}>
	<label>
		File
		<input type="file" accept="image/*" onchange={onFileChange} required />
	</label>
	<label>
		Alt text
		<input type="text" bind:value={altText} />
	</label>
	<label>
		File type
		<select bind:value={fileType}>
			<option value={0}>png</option>
			<option value={1}>jpg</option>
			<option value={2}>jpeg</option>
			<option value={3}>gif</option>
			<option value={4}>webp</option>
		</select>
	</label>
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create photo'}</button>
</form>

{#if error}
	<p class="error">{error}</p>
{/if}

<style>
	.error {
		color: #c00;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 24rem;
		margin-top: 1rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	input,
	select {
		padding: 0.35rem;
	}
</style>
