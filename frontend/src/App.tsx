import './index.css';
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./pages/login";
import Navbar from "./components/navbar";
import Dashboard from './pages/dashboard';

import { LeagueProvider } from './context/LeagueContext';
import Draftboard from './pages/draftboard';
import TeamSheets from './pages/teamsheets';
import DraftHistory from './pages/DraftHistory';

function App() {
    return (
        <BrowserRouter>
            <Routes>
                {/* --- Public routes --- */}
                <Route path='/' element={<Login />} /> {/* placeholder */}
                <Route path='/login' element={<Login />} />
                <Route path='/faq' element={<Login />} /> {/* placeholder */}

                {/* --- Protected Global Routes (user must be logged in) --- */}
                {/* this route would list all leagues the user is a part off */}
                <Route path='/my-leagues' element={null} />

                {/* --- Admin User Routes (Protected by the global 'admin' role) --- */}
                {/* not implemented */}
                {/* <Route path='/admin/*' element={<AdminRouteGuard><AdminDashboard></AdminRouteGuard>} /> */}

                {/* --- Protected League-Specific Routes (Requires league context) --- */}
                <Route path='/league/:leagueId/*' element={<LeagueProviderWrapper />} />
            </Routes>
        </BrowserRouter>
    );
}

// Wrapper component to use LeagueProvider and define league-specific routes
function LeagueProviderWrapper() {
    return (
        <LeagueProvider> {/* extracts leagueId from URL*/}
            <Navbar />
            <Routes>
                <Route path='dashboard' element={<Dashboard />} />
                <Route path='draftboard' element={<Draftboard />} />
                <Route path='teamsheets' element={<TeamSheets />} />
                <Route path='draft-history' element={<DraftHistory />} />
                {/* <Route path='standings' element={<Standings />} /> */}

                {/* --- League Staff Routes (Protected by league-specific roles) */}
                {/* placeholder examples */}

                {/* <Route path='staff/edit-rules' element={<LeagueStaffRouteGuard role="owner"><EditRules /></LeagueStaffRouteGuard>} /> */}

            </Routes>
        </LeagueProvider>
    );
}

export default App;
