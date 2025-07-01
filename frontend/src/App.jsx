import './index.css'
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./pages/login"
import Navbar from "./componenets/navbar"
import Draftboard from "./pages/draftboard"
function App() {

  return (
    <BrowserRouter>
      <Routes>
        <Route path='/' element={<Login />}/>
        <Route path='/bruh' element={<Draftboard />}/>
      </Routes>
    </BrowserRouter>
  )
}

export default App
