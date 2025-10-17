import { useEffect, useState } from 'react'
import './styles/theme.css'
import './App.css'
import Login from './components/Login'
import Dashboard from './components/Dashboard'
import UserManagement from './components/UserManagement'
import Groups from './components/Groups'
import Layout from './components/Layout'

function App() {
  const [currentPath, setCurrentPath] = useState(window.location.pathname)

  useEffect(() => {
    const handleLocationChange = () => {
      setCurrentPath(window.location.pathname)
    }

    // Listen for browser navigation
    window.addEventListener('popstate', handleLocationChange)

    return () => {
      window.removeEventListener('popstate', handleLocationChange)
    }
  }, [])

  // Check if user is authenticated
  const token = localStorage.getItem('token') || sessionStorage.getItem('token')

  // Protected routes - require authentication
  const protectedRoutes = ['/dashboard', '/users', '/groups']
  const isProtectedRoute = protectedRoutes.some(route => currentPath.startsWith(route))

  // If on protected route but not authenticated, redirect to login
  if (isProtectedRoute && !token) {
    window.location.href = '/'
    return null
  }

  // If authenticated and on root, redirect to dashboard
  if (currentPath === '/' && token) {
    window.location.href = '/dashboard'
    return null
  }

  // Route handling with Layout
  if (isProtectedRoute) {
    return (
      <Layout currentPath={currentPath}>
        {currentPath === '/dashboard' && <Dashboard />}
        {currentPath === '/users' && <UserManagement />}
        {currentPath === '/groups' && <Groups />}
      </Layout>
    )
  }

  // Default to login
  return <Login />
}

export default App
