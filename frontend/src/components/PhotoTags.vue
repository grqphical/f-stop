<script setup lang="ts">
import { ref } from 'vue';
import type { Tag } from '../models/models';
import { PhTrash } from '@phosphor-icons/vue';

const props = defineProps<{
    photoId: string;
    tags: Tag[];
}>();

const emit = defineEmits<{
    (e: 'removed', photoId: string, tagId: number): void;
}>();

const confirmingTag = ref<Tag | null>(null);
const deletingId = ref<number | null>(null);
const error = ref<string | null>(null);

function requestRemove(tag: Tag): void {
    confirmingTag.value = tag;
    error.value = null;
}

function cancelRemove(): void {
    confirmingTag.value = null;
    error.value = null;
}

async function confirmRemove(): Promise<void> {
    const tag = confirmingTag.value;
    if (!tag) return;
    deletingId.value = tag.id;
    error.value = null;
    try {
        const response = await fetch(`/api/v1/photo/${props.photoId}/tags`, {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify({ tags: [tag.id] }),
        });
        if (!response.ok) {
            throw new Error(`Remove failed with status ${response.status}`);
        }
        emit('removed', props.photoId, tag.id);
        confirmingTag.value = null;
    } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to remove tag';
    } finally {
        deletingId.value = null;
    }
}
</script>

<template>
    <div class="contents">
        <button v-for="tag in tags" :key="tag.id" @click="requestRemove(tag)" :title="`Remove tag ${tag.name}`"
            aria-label="Remove tag"
            class="group inline-flex cursor-pointer items-center justify-center rounded-md bg-gray-300 px-1 text-center hover:bg-red-300 min-w-8">
            <span class="group-hover:hidden">{{ tag.name }}</span>
            <PhTrash size="12px" weight="bold" class="hidden group-hover:block" />
        </button>

        <div v-if="confirmingTag" @click.self="cancelRemove"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div class="w-full max-w-sm rounded-lg bg-white p-4 shadow-xl" role="dialog" aria-modal="true"
                aria-label="Confirm remove tag">
                <h2 class="text-base font-semibold text-slate-900">Remove tag?</h2>
                <p class="mt-1 text-sm text-slate-600">
                    Remove tag <strong>{{ confirmingTag.name }}</strong> from this photo?
                </p>
                <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>
                <div class="mt-4 flex justify-end gap-2">
                    <button @click="cancelRemove"
                        class="cursor-pointer rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50">
                        Cancel
                    </button>
                    <button @click="confirmRemove" :disabled="deletingId !== null"
                        class="cursor-pointer rounded-md bg-red-600 px-3 py-1.5 text-sm text-white hover:bg-red-500 disabled:cursor-not-allowed disabled:opacity-50">
                        {{ deletingId !== null ? 'Removing...' : 'Remove' }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
