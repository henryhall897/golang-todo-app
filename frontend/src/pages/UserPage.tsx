import { useAuth0, LogoutOptions } from '@auth0/auth0-react'

export default function UserPage() {
  const { user, getAccessTokenSilently, isAuthenticated, logout } = useAuth0()

  const handleProtectedCall = async () => {
    try {
      const token = await getAccessTokenSilently()
      console.log('Access Token:', token)

      const res = await fetch('http://localhost:8080/users', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      const data = await res.json()
      console.log(data)
    } catch (err) {
      console.error('API request failed:', err)
    }
  }

  return (
    <div>
      <h2>User Dashboard</h2>
      {isAuthenticated && <p>Logged in as: {user?.email}</p>}
      <button onClick={handleProtectedCall} className="mt-4 bg-blue-500 text-white px-4 py-2 rounded">
        Call Protected API
      </button>
      <button onClick={() => logout({ returnTo: window.location.origin } as LogoutOptions)} className="mt-2 bg-red-500 text-white px-4 py-2 rounded">
        Log Out
      </button>
    </div>
  )
}
