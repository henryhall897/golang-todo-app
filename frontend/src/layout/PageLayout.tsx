import Header from '../components/Header'
import Footer from '../components/Footer'

type PageLayoutProps = {
  children: React.ReactNode
}

export default function PageLayout({ children }: PageLayoutProps) {
  return (
    <div className="flex flex-col min-h-screen bg-gray-50 text-gray-900">
      <Header />
      <main className="flex-grow p-4 max-w-4xl mx-auto">{children}</main>
      <Footer />
    </div>
  )
}
