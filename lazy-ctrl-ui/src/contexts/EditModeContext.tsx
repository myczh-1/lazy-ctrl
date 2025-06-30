import { createContext, useContext, useState, type ReactNode } from 'react'

interface EditModeContextType {
  editMode: boolean
  setEditMode: (mode: boolean) => void
  toggleEditMode: () => void
}

const EditModeContext = createContext<EditModeContextType | undefined>(undefined)

export function EditModeProvider({ children }: { children: ReactNode }) {
  const [editMode, setEditMode] = useState(false)

  const toggleEditMode = () => setEditMode(!editMode)

  return (
    <EditModeContext.Provider value={{ editMode, setEditMode, toggleEditMode }}>
      {children}
    </EditModeContext.Provider>
  )
}

export function useEditMode() {
  const context = useContext(EditModeContext)
  if (context === undefined) {
    throw new Error('useEditMode must be used within an EditModeProvider')
  }
  return context
}