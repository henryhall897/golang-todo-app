import { Link } from 'react-router-dom'

export default function Header() {
  return (
    <header className="bg-blue-600 text-white shadow">
      <div className="max-w-6xl mx-auto px-4 py-3 flex justify-between items-center">
        <h1 className="text-lg font-semibold">Golang Todo App</h1>
        <nav className="space-x-4">
          <Link to="/" className="hover:underline">Home</Link>
          <Link to="/user" className="hover:underline">User</Link>
          <Link to="/admin" className="hover:underline">Admin</Link>
        </nav>
      </div>
    </header>
  )
}
