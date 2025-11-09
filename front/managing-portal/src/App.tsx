import { useEffect, useState } from 'react'
import './styles/theme.css'
import './App.css'
import Login from './components/Login'
import Dashboard from './components/Dashboard'
import UserManagement from './components/UserManagement'
import Groups from './components/Groups'
import Departments from './components/Departments'
import MeetingSubjects from './components/MeetingSubjects'
import Rooms from './components/Rooms'
import RoomDetails from './components/RoomDetails'
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
  const protectedRoutes = ['/dashboard', '/users', '/groups', '/departments', '/subjects', '/rooms']
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
    // Extract room SID from path if on room details page
    const roomMatch = currentPath.match(/^\/rooms\/([^/]+)$/);
    const roomSid = roomMatch ? roomMatch[1] : null;

    return (
      <Layout currentPath={currentPath}>
        {currentPath === '/dashboard' && <Dashboard />}
        {currentPath === '/users' && <UserManagement />}
        {currentPath === '/groups' && <Groups />}
        {currentPath === '/departments' && <Departments />}
        {currentPath === '/subjects' && <MeetingSubjects />}
        {currentPath === '/rooms' && <Rooms />}
        {roomSid && <RoomDetails roomSid={roomSid} />}
      </Layout>
    )
  }

  // Default to login
  return <Login />
}

export default App
