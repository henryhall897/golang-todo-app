import { useAuth0 } from '@auth0/auth0-react'
import PageLayout from '../layout/PageLayout'

export default function LoginPage() {
  const { loginWithRedirect, isAuthenticated, isLoading, error } = useAuth0()
  if (isLoading) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>
  }
  if (error) { 
    return <div className="flex items-center justify-center h-screen">Error: {error.message}</div>
  }

  return (
    <PageLayout>
      <div className="max-w-md mx-auto bg-white shadow-md rounded p-6 space-y-4">
        <h1 className="text-2xl font-bold text-center text-blue-600">Welcome to My Golang Todo App</h1>

        {!isAuthenticated && (
          <button
            onClick={() => loginWithRedirect()}
            className="w-full bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2 px-4 rounded transition duration-200"
          >
            Log in to Continue
          </button>
        )}
      </div>
    </PageLayout>
  )
}
