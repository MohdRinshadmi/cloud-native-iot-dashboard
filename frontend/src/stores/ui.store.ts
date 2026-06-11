import { create } from 'zustand';
import { persist } from 'zustand/middleware';

/**
 * Global UI chrome state. Kept deliberately tiny — page/server state belongs
 * to TanStack Query, form state to React Hook Form. Zustand only holds what
 * is truly global and client-owned.
 */
interface UIState {
  sidebarCollapsed: boolean;
  commandPaletteOpen: boolean;
  toggleSidebar: () => void;
  setCommandPaletteOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarCollapsed: false,
      commandPaletteOpen: false,
      toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
      setCommandPaletteOpen: (open) => set({ commandPaletteOpen: open }),
    }),
    {
      name: 'iot-ui',
      // The palette is ephemeral; only layout prefs survive a reload.
      partialize: (s) => ({ sidebarCollapsed: s.sidebarCollapsed }),
    },
  ),
);
