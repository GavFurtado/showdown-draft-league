import './index.css';
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import Login from "./pages/login";
import Navbar from "./components/navbar";
import Dashboard from './pages/dashboard';

import { LeagueProvider } from './context/LeagueContext';
import { UserProvider } from './context/UserContext';
import Draftboard from './pages/draftboard';
import TeamSheets from './pages/teamsheets';
import DraftHistory from './pages/DraftHistory';
import JoinLeague from './pages/JoinLeague';
import { DraftHistoryProvider } from './context/DraftHistoryContext';

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
                <Route element={<UserProviderWrapper />}>
                    <Route path='/my-leagues' element={null} />
                    <Route path='/:leagueId/join' element={<JoinLeague />} />
                </Route>

                {/* --- Admin User Routes (Protected by the global 'admin' role) --- */}
                {/* not implemented */}
                {/* <Route path='/admin/*' element={<AdminRouteGuard><AdminDashboard></AdminRouteGuard>} /> */}

                {/* --- Protected League-Specific Routes (Requires league context) --- */}
                <Route path='/league/:leagueId/*' element={<UserAndLeagueProviderWrapper />} />
            </Routes>
        </BrowserRouter>
    );
}

function UserProviderWrapper() {
    return (
        <UserProvider>
            <Navbar />
            <Outlet />
        </UserProvider>
    );
}

function UserAndLeagueProviderWrapper() {
    return (
        <UserProvider>
            <LeagueProvider> {/* extracts leagueId from URL*/}
                <DraftHistoryProvider>
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
                </DraftHistoryProvider>
            </LeagueProvider>
        </UserProvider>
    );
}



export default App;
