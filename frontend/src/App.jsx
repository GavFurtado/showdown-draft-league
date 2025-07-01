import './index.css'
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./pages/login"
import Navbar from "./componenets/navbar"
function App() {

  return (
    <BrowserRouter>
      <Routes>
        <Route path='/' element={<Navbar />}/>
      </Routes>
    </BrowserRouter>
  )
}

export default App
