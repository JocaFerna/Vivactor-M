import { create } from 'zustand'

export const useGlobalStore = create((set) => ({
    architectureURL: null,
}))