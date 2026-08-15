<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getSeason } from '$lib/season';
	import type { Season } from '$lib/season';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { getCurrentUserId } from '$lib/auth';
	import {
		REACTIONS,
		reactionByValue,
		listBlurbs,
		createBlurb,
		deleteBlurb,
		listBlurbPhotos,
		listBlurbReactions,
		reactToBlurb,
		unreactToBlurb,
		addBlurbPhoto,
		removeBlurbPhoto
	} from '$lib/blurb';
	import type { Blurb } from '$lib/blurb';
	import {
		listComments,
		createComment,
		deleteComment,
		listCommentReactions,
		reactToComment,
		unreactToComment
	} from '$lib/comment';
	import type { Comment } from '$lib/comment';
	import { listPhotos, createPhoto, dataUrlFor, photoTypeFromExtension } from '$lib/photo';
	import type { Photo, PhotoType } from '$lib/photo';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

	const id = () => page.params.id as string;

	let season = $state<Season | null>(null);
	let users = $state<User[]>([]);
	let blurbs = $state<Blurb[]>([]);
	let comments = $state<Comment[]>([]);
	let photosById = $state<Record<string, Photo>>({});
	let blurbPhotosByBlurb = $state<Record<string, Photo[]>>({});

	// create-blurb form state
	let newTitle = $state('');
	let newContent = $state('');
	let newPhotoContents = $state('');
	let newPhotoFileType = $state<PhotoType>(0);
	let newPhotoAlt = $state('');
	let newPhotoReady = $state(false);

	// per-blurb comment form state
	let commentDrafts = $state<Record<string, string>>({});
	let replyDrafts = $state<Record<string, string>>({});
	let replyingTo = $state<Record<string, string>>({});

	let loading = $state(true);
	let loadError = $state('');
	let formError = $state('');

	const currentUserId = $derived(getCurrentUserId() ?? '');

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	const blurbsForSeason = $derived(blurbs.filter((b) => b.season === id()));

	// photos attached to each blurb
	function photosFor(blurbId: string): Photo[] {
		return blurbPhotosByBlurb[blurbId] ?? [];
	}

	// aggregate reactions per blurb, with counts and whether the current user reacted
	function reactionSummary(
		rows: { user_id: string; reaction_type: number }[]
	): { value: number; name: string; emoji: string; count: number; mine: boolean }[] {
		const byValue: Record<number, { count: number; mine: boolean }> = {};
		for (const r of rows) {
			const e = byValue[r.reaction_type] ?? { count: 0, mine: false };
			e.count += 1;
			if (r.user_id === currentUserId) e.mine = true;
			byValue[r.reaction_type] = e;
		}
		return REACTIONS.map((rt) => ({
			value: rt.value,
			name: rt.name,
			emoji: rt.emoji,
			count: byValue[rt.value]?.count ?? 0,
			mine: byValue[rt.value]?.mine ?? false
		})).filter((s) => s.count > 0);
	}

	// reaction rows per blurb / comment, populated by load(); $state so the
	// mutations in load() are reactive (a $derived value must not be mutated)
	let blurbReactionsByBlurb = $state<Record<string, { user_id: string; reaction_type: number }[]>>({});
	let commentReactionsByComment = $state<Record<string, { user_id: string; reaction_type: number }[]>>({});

	// whether the current user already reacted with the given named reaction
	function blurbReactionMine(blurbId: string, name: string): boolean {
		return reactionSummary(blurbReactionsByBlurb[blurbId] ?? []).find((s) => s.name === name)?.mine ?? false;
	}

	function commentReactionMine(commentId: string, name: string): boolean {
		return reactionSummary(commentReactionsByComment[commentId] ?? []).find((s) => s.name === name)?.mine ?? false;
	}

	function commentsFor(blurbId: string): Comment[] {
		return comments.filter((c) => c.blurb === blurbId);
	}

	// top-level comments (no reply_to), oldest first
	function topLevelComments(blurbId: string): Comment[] {
		return commentsFor(blurbId)
			.filter((c) => !c.reply_to || c.reply_to === '')
			.sort((a, b) => (a.created_at < b.created_at ? -1 : 1));
	}

	// replies to a comment, oldest first
	function repliesTo(blurbId: string, commentId: string): Comment[] {
		return commentsFor(blurbId)
			.filter((c) => c.reply_to === commentId)
			.sort((a, b) => (a.created_at < b.created_at ? -1 : 1));
	}

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [seasonData, userList, blurbList, commentList, blurbPhotoList, blurbReactionList, commentReactionList, photoList] =
				await Promise.all([
					getSeason(id()),
					listUsers(),
					listBlurbs(),
					listComments(),
					listBlurbPhotos(),
					listBlurbReactions(),
					listCommentReactions(),
					listPhotos()
				]);
			season = seasonData;
			users = userList;
			blurbs = blurbList;
			comments = commentList;
			photosById = Object.fromEntries(photoList.map((p) => [p.id, p]));

			for (const blurb of blurbList) {
				blurbReactionsByBlurb[blurb.id] = blurbReactionList.filter((r) => r.blurb_id === blurb.id);
			}
			for (const comment of commentList) {
				commentReactionsByComment[comment.id] = commentReactionList.filter((r) => r.comment_id === comment.id);
			}
			// map blurb -> attached photos
			for (const blurb of blurbList) {
				blurbPhotosByBlurb[blurb.id] = blurbPhotoList
					.filter((p) => p.blurb_id === blurb.id)
					.map((p) => photoList.find((ph) => ph.id === p.photo_id))
					.filter((ph): ph is Photo => Boolean(ph));
			}
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load blurbs';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function onPhotoFileChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		const type = photoTypeFromExtension(file.name);
		if (type === null) {
			formError = 'Unsupported file type. Use png, jpg, jpeg, gif, or webp.';
			newPhotoReady = false;
			return;
		}
		newPhotoFileType = type;
		newPhotoReady = false;
		formError = '';
		const reader = new FileReader();
		reader.onload = () => {
			newPhotoContents = (reader.result as string).split(',')[1] ?? '';
			newPhotoReady = true;
		};
		reader.readAsDataURL(file);
	}

	async function handleCreateBlurb(e: Event) {
		e.preventDefault();
		formError = '';
		if (!newTitle.trim() || !newContent.trim()) {
			formError = 'Title and content are required.';
			return;
		}
		try {
			const created = await createBlurb({ title: newTitle.trim(), content: newContent.trim(), season: id() });
			// optionally attach an uploaded photo
			if (newPhotoContents) {
				const photo = await createPhoto({ alt_text: newPhotoAlt, contents: newPhotoContents, file_type: newPhotoFileType });
				await addBlurbPhoto(created.id, photo.id);
			}
			newTitle = '';
			newContent = '';
			newPhotoContents = '';
			newPhotoAlt = '';
			newPhotoReady = false;
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to create blurb';
		}
	}

	async function handleDeleteBlurb(blurbId: string) {
		if (!window.confirm('Delete this blurb?')) return;
		try {
			await deleteBlurb(blurbId);
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to delete blurb';
		}
	}

	async function toggleBlurbReaction(blurbId: string, name: string, mine: boolean) {
		try {
			if (mine) {
				await unreactToBlurb(blurbId, name);
			} else {
				await reactToBlurb(blurbId, name);
			}
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to update reaction';
		}
	}

	async function handleAddComment(blurbId: string) {
		const content = (commentDrafts[blurbId] ?? '').trim();
		if (!content) return;
		try {
			await createComment(blurbId, content);
			commentDrafts[blurbId] = '';
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to add comment';
		}
	}

	function startReply(commentId: string) {
		replyingTo[commentId] = replyingTo[commentId] ? '' : commentId;
		formError = '';
	}

	async function handleReply(commentId: string) {
		const content = (replyDrafts[commentId] ?? '').trim();
		if (!content) return;
		const parent = comments.find((c) => c.id === commentId);
		if (!parent) return;
		try {
			await createComment(parent.blurb, content, parent.id);
			replyDrafts[commentId] = '';
			replyingTo[commentId] = '';
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to reply';
		}
	}

	async function handleDeleteComment(commentId: string) {
		if (!window.confirm('Delete this comment?')) return;
		try {
			await deleteComment(commentId);
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to delete comment';
		}
	}

	async function toggleCommentReaction(commentId: string, name: string, mine: boolean) {
		try {
			if (mine) {
				await unreactToComment(commentId, name);
			} else {
				await reactToComment(commentId, name);
			}
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to update reaction';
		}
	}

	async function handleRemovePhoto(blurbId: string, photoId: string) {
		try {
			await removeBlurbPhoto(blurbId, photoId);
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : 'Failed to remove photo';
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {season ? season.name : 'Season'} — Blurbs</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">{season?.name ?? 'Season'} — Blurbs</h1>
	<div class="ml-auto">
		<a href={`/seasons/${id()}`} class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to season</a>
	</div>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else}
	{#if formError}
		<p class="mt-4 text-sm font-medium text-destructive">{formError}</p>
	{/if}

	<!-- Create blurb -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle>Post a blurb</CardTitle>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleCreateBlurb} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="title">Title</Label>
					<Input id="title" type="text" bind:value={newTitle} placeholder="e.g. Season kickoff" />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="content">Content</Label>
					<Textarea id="content" bind:value={newContent} placeholder="What's happening?" />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="photo">Attach a photo (optional)</Label>
					<Input id="photo" type="file" accept="image/*" onchange={onPhotoFileChange} />
					{#if newPhotoReady}
						<p class="text-sm text-muted-foreground">Photo ready to attach</p>
					{/if}
				</div>
				<Button type="submit" class="w-fit">Post blurb</Button>
			</form>
		</CardContent>
	</Card>

	{#if blurbsForSeason.length === 0}
		<p class="mt-6 text-muted-foreground">No blurbs yet for this season.</p>
	{/if}

	<!-- Feed -->
	<div class="mt-6 flex flex-col gap-6">
		{#each blurbsForSeason as blurb (blurb.id)}
			<Card>
				<CardHeader>
					<div class="flex items-center gap-3">
						<CardTitle>{blurb.title}</CardTitle>
						<div class="ml-auto flex items-center gap-2">
							<Badge variant="secondary">{userName(blurb.owner)}</Badge>
							{#if blurb.owner === currentUserId}
								<Button size="sm" variant="ghost" onclick={() => handleDeleteBlurb(blurb.id)}>Delete</Button>
							{/if}
						</div>
					</div>
				</CardHeader>
				<CardContent class="flex flex-col gap-4">
					<p class="text-sm whitespace-pre-wrap">{blurb.content}</p>

					{#if photosFor(blurb.id).length > 0}
						<div class="flex flex-wrap gap-3">
							{#each photosFor(blurb.id) as photo}
								<div class="relative">
									<img
										src={dataUrlFor(photo)}
										alt={photo.alt_text || blurb.title}
										class="h-32 w-48 rounded-md border border-border object-cover"
									/>
									{#if blurb.owner === currentUserId}
										<button
											class="absolute right-1 top-1 rounded-full bg-background/80 px-1.5 text-xs text-muted-foreground hover:text-destructive"
											onclick={() => handleRemovePhoto(blurb.id, photo.id)}
											aria-label="Remove photo"
										>
											✕
										</button>
									{/if}
								</div>
							{/each}
						</div>
					{/if}

					<!-- Reactions -->
					{#if reactionSummary(blurbReactionsByBlurb[blurb.id] ?? []).length > 0}
						<div class="flex flex-wrap gap-2">
							{#each reactionSummary(blurbReactionsByBlurb[blurb.id] ?? []) as r}
								<Button
									size="sm"
									variant={r.mine ? 'default' : 'secondary'}
									onclick={() => toggleBlurbReaction(blurb.id, r.name, r.mine)}
								>
									{r.emoji} {r.count}
								</Button>
							{/each}
						</div>
					{/if}
					<div class="flex flex-wrap gap-2">
						{#each REACTIONS as rt}
							<Button
								size="sm"
								variant="ghost"
								title={rt.name}
								onclick={() => toggleBlurbReaction(blurb.id, rt.name, blurbReactionMine(blurb.id, rt.name))}
							>
								{rt.emoji}
							</Button>
						{/each}
					</div>

					<!-- Comments -->
					<div class="mt-2 flex flex-col gap-3">
						{#each topLevelComments(blurb.id) as comment}
							<div class="rounded-lg border border-border p-3">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium">{userName(comment.user_id)}</span>
									<span class="text-xs text-muted-foreground">{new Date(comment.created_at).toLocaleString()}</span>
									<div class="ml-auto flex items-center gap-2">
										{#if reactionSummary(commentReactionsByComment[comment.id] ?? []).length > 0}
											{#each reactionSummary(commentReactionsByComment[comment.id] ?? []) as r}
												<Button
													size="sm"
													variant={r.mine ? 'default' : 'secondary'}
													onclick={() => toggleCommentReaction(comment.id, r.name, r.mine)}
												>
													{r.emoji} {r.count}
												</Button>
											{/each}
										{/if}
										<Button size="sm" variant="ghost" onclick={() => startReply(comment.id)}>Reply</Button>
										{#if comment.user_id === currentUserId}
											<Button size="sm" variant="ghost" onclick={() => handleDeleteComment(comment.id)}>Delete</Button>
										{/if}
									</div>
								</div>
								<p class="mt-1 text-sm whitespace-pre-wrap">{comment.content}</p>

								<div class="mt-1 flex flex-wrap items-center gap-1">
									{#each REACTIONS as rt}
										<Button
											size="sm"
											variant="ghost"
											title={rt.name}
											onclick={() => toggleCommentReaction(comment.id, rt.name, commentReactionMine(comment.id, rt.name))}
										>
											{rt.emoji}
										</Button>
									{/each}
								</div>

								{#if replyingTo[comment.id]}
									<div class="mt-2 flex gap-2">
										<Input
											placeholder="Write a reply..."
											bind:value={replyDrafts[comment.id]}
											onkeydown={(e) => e.key === 'Enter' && handleReply(comment.id)}
										/>
										<Button size="sm" onclick={() => handleReply(comment.id)}>Reply</Button>
									</div>
								{/if}

								{#if repliesTo(blurb.id, comment.id).length > 0}
									<div class="mt-2 flex flex-col gap-2 border-l-2 border-border pl-3">
										{#each repliesTo(blurb.id, comment.id) as reply}
											<div class="flex items-start justify-between gap-2">
												<div>
													<span class="text-sm font-medium">{userName(reply.user_id)}</span>
													<span class="text-xs text-muted-foreground"> · {new Date(reply.created_at).toLocaleString()}</span>
													<p class="text-sm whitespace-pre-wrap">{reply.content}</p>
													<div class="mt-1 flex flex-wrap items-center gap-1">
														{#each REACTIONS as rt}
															<Button
																size="sm"
																variant="ghost"
																title={rt.name}
																onclick={() => toggleCommentReaction(reply.id, rt.name, commentReactionMine(reply.id, rt.name))}
															>
																{rt.emoji}
															</Button>
														{/each}
													</div>
												</div>
												<div class="flex items-center gap-1">
													{#each reactionSummary(commentReactionsByComment[reply.id] ?? []) as r}
														<Button
															size="sm"
															variant={r.mine ? 'default' : 'secondary'}
															onclick={() => toggleCommentReaction(reply.id, r.name, r.mine)}
														>
															{r.emoji} {r.count}
														</Button>
													{/each}
													{#if reply.user_id === currentUserId}
														<Button size="sm" variant="ghost" onclick={() => handleDeleteComment(reply.id)}>Delete</Button>
													{/if}
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						{/each}

						<div class="flex gap-2">
							<Input
								placeholder="Add a comment..."
								bind:value={commentDrafts[blurb.id]}
								onkeydown={(e) => e.key === 'Enter' && handleAddComment(blurb.id)}
							/>
							<Button size="sm" onclick={() => handleAddComment(blurb.id)}>Comment</Button>
						</div>
					</div>
				</CardContent>
			</Card>
		{/each}
	</div>
{/if}
