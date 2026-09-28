<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { PhotoMetadata, Tag } from '../models/models';
import { PhX } from '@phosphor-icons/vue';

const props = defineProps<{
    photo: PhotoMetadata | null;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'assigned', photoId: string, tags: Tag[]): void;
}>();

const allTags = ref<Tag[]>([]);
const selectedIds = ref<Set<number>>(new Set());
const loadingTags = ref(false);
const saving = ref(false);
const creating = ref(false);
const newTagName = ref('');
const error = ref<string | null>(null);

const assignedIds = computed(() => new Set((props.photo?.tags ?? []).map(t => t.id)));

const availableTags = computed(() => allTags.value);
const selectedCount = computed(() => selectedIds.value.size);

function reset() {
    allTags.value = [];
    selectedIds.value = new Set();
    newTagName.value = '';
    error.value = null;
    saving.value = false;
    creating.value = false;
    loadingTags.value = false;
}

watch(() => props.photo, async (photo) => {
    reset();
    if (!photo) return;
    await fetchTags();
}, { immediate: true });

async function fetchTags(): Promise<void> {
    loadingTags.value = true;
    error.value = null;
    try {
        const response = await fetch('/api/v1/tags', { credentials: 'same-origin' });
        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }
        const json = await response.json() as { tags: Tag[] };
        allTags.value = json.tags ?? [];
    } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to load tags';
    } finally {
        loadingTags.value = false;
    }
}

function toggleSelect(tagId: number): void {
    if (assignedIds.value.has(tagId)) return;
    const next = new Set(selectedIds.value);
    if (next.has(tagId)) {
        next.delete(tagId);
    } else {
        next.add(tagId);
    }
    selectedIds.value = next;
}

async function assignSelected(): Promise<void> {
    if (!props.photo || selectedIds.value.size === 0) return;
    saving.value = true;
    error.value = null;
    try {
        const ids = [...selectedIds.value];
        const response = await fetch(`/api/v1/photo/${props.photo.id}/tags`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify({ tags: ids }),
        });
        if (!response.ok) {
            throw new Error(`Assign failed with status ${response.status}`);
        }
        const assigned = allTags.value.filter(t => ids.includes(t.id));
        emit('assigned', props.photo.id, assigned);
        emit('close');
    } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to assign tags';
    } finally {
        saving.value = false;
    }
}

async function createAndAssign(): Promise<void> {
    if (!props.photo) return;
    const name = newTagName.value.trim();
    if (!name) return;
    creating.value = true;
    error.value = null;
    try {
        const formData = new FormData();
        formData.append('name', name);
        const createRes = await fetch('/api/v1/tags', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin',
        });
        if (!createRes.ok) {
            const body = await createRes.text().catch(() => '');
            throw new Error(
                createRes.status === 409
                    ? `A tag named "${name}" already exists`
                    : `Create tag failed with status ${createRes.status}${body ? `: ${body}` : ''}`
            );
        }
        const { tagId } = await createRes.json() as { tagId: number };

        const assignRes = await fetch(`/api/v1/photo/${props.photo.id}/tags`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify({ tags: [tagId] }),
        });
        if (!assignRes.ok) {
            throw new Error(`Assign failed with status ${assignRes.status}`);
        }

        const newTag: Tag = { id: tagId, ownerID: props.photo.ownerID, name };
        allTags.value = [...allTags.value, newTag];
        emit('assigned', props.photo.id, [newTag]);
        newTagName.value = '';
        emit('close');
    } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to create tag';
    } finally {
        creating.value = false;
    }
}

function onBackdropClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) emit('close');
}
</script>

<template>
    <div v-if="photo" @click="onBackdropClick"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
        <div class="w-full max-w-md rounded-lg bg-white shadow-xl" role="dialog" aria-modal="true"
            aria-label="Assign tags">
            <div class="flex items-center justify-between border-b border-slate-200 px-4 py-3">
                <h2 class="text-base font-semibold text-slate-900">Add tags</h2>
                <button @click="emit('close')" class="cursor-pointer rounded p-1 text-slate-500 hover:bg-slate-100"
                    aria-label="Close">
                    <PhX size="16px" weight="bold" />
                </button>
            </div>

            <div class="px-4 py-3">
                <p v-if="error" class="mb-2 text-sm text-red-600">{{ error }}</p>

                <p class="mb-2 text-sm font-medium text-slate-700">Select existing tags</p>
                <div v-if="loadingTags" class="py-4 text-sm text-slate-500">Loading tags...</div>
                <div v-else-if="availableTags.length === 0" class="py-4 text-sm text-slate-500">
                    No tags yet. Create one below.
                </div>
                <ul v-else class="max-h-56 divide-y divide-slate-100 overflow-y-auto rounded-md border border-slate-200">
                    <li v-for="tag in availableTags" :key="tag.id">
                        <label class="flex cursor-pointer items-center gap-2 px-3 py-2 hover:bg-slate-50"
                            :class="{ 'opacity-60': assignedIds.has(tag.id) }">
                            <input type="checkbox" :checked="assignedIds.has(tag.id) || selectedIds.has(tag.id)"
                                :disabled="assignedIds.has(tag.id)" @change="toggleSelect(tag.id)"
                                class="h-4 w-4 accent-indigo-600" />
                            <span class="text-sm text-slate-800">{{ tag.name }}</span>
                            <span v-if="assignedIds.has(tag.id)"
                                class="ml-auto text-xs text-slate-400">already added</span>
                        </label>
                    </li>
                </ul>

                <div class="mt-4">
                    <p class="mb-2 text-sm font-medium text-slate-700">Or create a new tag</p>
                    <div class="flex gap-2">
                        <input v-model="newTagName" type="text" placeholder="New tag name" maxlength="64"
                            @keyup.enter="createAndAssign"
                            class="min-w-0 flex-1 rounded-md border border-slate-300 px-2 py-1.5 text-sm focus:border-indigo-500 focus:outline-none" />
                        <button @click="createAndAssign" :disabled="!newTagName.trim() || creating"
                            class="shrink-0 cursor-pointer rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50">
                            {{ creating ? 'Creating...' : 'Create & assign' }}
                        </button>
                    </div>
                </div>
            </div>

            <div class="flex justify-end gap-2 border-t border-slate-200 px-4 py-3">
                <button @click="emit('close')"
                    class="cursor-pointer rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50">
                    Cancel
                </button>
                <button @click="assignSelected" :disabled="selectedCount === 0 || saving"
                    class="cursor-pointer rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50">
                    {{ saving ? 'Assigning...' : `Assign${selectedCount > 0 ? ` (${selectedCount})` : ''}` }}
                </button>
            </div>
        </div>
    </div>
</template>
