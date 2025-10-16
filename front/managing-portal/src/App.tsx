import { useEffect, useState } from 'react'
import './styles/theme.css'
import './App.css'
import Login from './components/Login'
import Dashboard from './components/Dashboard'

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

  // If on dashboard route but not authenticated, redirect to login
  if (currentPath === '/dashboard' && !token) {
    window.location.href = '/'
    return null
  }

  // If authenticated and on root, redirect to dashboard
  if (currentPath === '/' && token) {
    window.location.href = '/dashboard'
    return null
  }

  // Show dashboard if on /dashboard route
  if (currentPath === '/dashboard') {
    return <Dashboard />
  }

  // Default to login
  return <Login />
}

export default App
