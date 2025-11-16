import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import './styles/theme.css';
import './App.css';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import Meetings from './pages/Meetings';
import MeetingRoom from './pages/MeetingRoom';
import MeetingRecordings from './pages/MeetingRecordings';
import Search from './pages/Search';
import Documents from './pages/Documents';
import Management from './pages/Management';
import Profile from './pages/Profile';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';
import AnonymousJoin from './pages/AnonymousJoin';
import Forbidden from './pages/Forbidden';
import MeetingDisconnected from './pages/MeetingDisconnected';

// Component to redirect to meetings if logged in, otherwise to login
function RootRedirect() {
  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
  return token ? <Navigate to="/dashboard/meetings" replace /> : <Navigate to="/login" replace />;
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<RootRedirect />} />
        <Route path="/login" element={<Login />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/join/:meetingId" element={<AnonymousJoin />} />
        <Route path="/meeting/:meetingId/join" element={<AnonymousJoin />} />
        <Route path="/forbidden" element={<Forbidden />} />
        <Route path="/dashboard" element={<Dashboard />}>
          <Route index element={<Navigate to="/dashboard/meetings" replace />} />
          <Route path="meetings" element={<Meetings />} />
          <Route path="search" element={<Search />} />
          <Route path="documents" element={<Documents />} />
          <Route path="management" element={<Management />} />
          <Route path="profile" element={<Profile />} />
        </Route>
        <Route path="/meeting/:meetingId" element={<MeetingRoom />} />
        <Route path="/meeting-room/:meetingId" element={<MeetingRoom />} />
        <Route path="/meeting/disconnected" element={<MeetingDisconnected />} />
        <Route path="/meeting/:meetingId/recordings" element={<MeetingRecordings />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
