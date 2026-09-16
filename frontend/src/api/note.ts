import request from '@/utils/request'
import type { NoteItem, TastingNote } from '@/constants/note'
import type { PageData } from '@/types/api'
import type { Comment } from '@/types/api'

export function listNotes(params: { page?: number; page_size?: number; roast?: string; origin?: string; keyword?: string; sort?: string }) {
  return request.get<never, PageData<NoteItem>>('/notes', { params })
}
export function getNote(id: number | string) { return request.get<never, { note: TastingNote; like_count: number }>(`/notes/${id}`) }
export function createNote(payload: Partial<TastingNote>) { return request.post<never, TastingNote>('/notes', payload) }
export function updateNote(id: number, payload: Partial<TastingNote>) { return request.put<never, TastingNote>(`/notes/${id}`, payload) }
export function deleteNote(id: number) { return request.delete<never, { deleted: boolean }>(`/notes/${id}`) }
export function listComments(noteId: number) { return request.get<never, Comment[]>(`/notes/${noteId}/comments`) }
export function createComment(noteId: number, content: string) { return request.post<never, Comment>(`/notes/${noteId}/comments`, { content }) }
export function likeNote(noteId: number) { return request.post<never, { id: number }>(`/notes/${noteId}/like`) }
export function unlikeNote(noteId: number) { return request.delete<never, { unliked: boolean }>(`/notes/${noteId}/like`) }
