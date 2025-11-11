import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import './styles/theme.css';
import './App.css';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import Meetings from './pages/Meetings';
import MeetingRoom from './pages/MeetingRoom';
import Search from './pages/Search';
import Documents from './pages/Documents';
import Management from './pages/Management';
import Profile from './pages/Profile';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/dashboard" element={<Dashboard />}>
          <Route index element={<Navigate to="/dashboard/meetings" replace />} />
          <Route path="meetings" element={<Meetings />} />
          <Route path="search" element={<Search />} />
          <Route path="documents" element={<Documents />} />
          <Route path="management" element={<Management />} />
          <Route path="profile" element={<Profile />} />
        </Route>
        <Route path="/meeting/:meetingId" element={<MeetingRoom />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
