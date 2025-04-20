import PageLayout from '../layout/PageLayout'

export default function AdminPage() {
  return (
    <PageLayout>
      <div className="space-y-4">
        <h1 className="text-3xl font-bold text-blue-700">Admin Dashboard</h1>
        <p className="text-gray-700">Welcome, admin! You have elevated access.</p>
        {/* Add admin-specific tools and data here later */}
      </div>
    </PageLayout>
  )
}
