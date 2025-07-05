import './index.css'
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./pages/login"
import Navbar from "./componenets/navbar"
import Dashboard from './pages/dashboard';
import StandingsTable from './pages/dashboard';
import Draftboard from "./pages/draftboard"

function App() {
  return (
    <BrowserRouter>
      { /* Add auth to Navbar so that it only shows when user is logged in */}
      <Routes>
        <Route path='/dashboard' element={<Dashboard />} />
        <Route path='/' element={<Login />} />
        {/* <Route path='/draftboard' element={<Draftboard />} /> */}
      </Routes>
    </BrowserRouter>
  )
}

export default App

