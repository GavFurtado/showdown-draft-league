import './index.css'
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./pages/login"
<<<<<<< HEAD
=======
import Navbar from "./componenets/navbar"
import Dashboard from './pages/dashboard';
import StandingsTable from './pages/dashboard';
>>>>>>> 27ecd243bf2545b0e3f8f32ef773f76c912435d5
import Draftboard from "./pages/draftboard"

function App() {
  return (
    <BrowserRouter>
      <Navbar />       { /* Add auth to Navbar so that it only shows when user is logged in */ }
      <Routes>
        <Route path='/dashboard' element={<Dashboard />}/>
        <Route path='/' element={<Login />} />
        <Route path='/bruh' element={<Draftboard />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
