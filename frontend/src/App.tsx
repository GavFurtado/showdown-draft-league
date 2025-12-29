import './index.css';
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import Navbar from "./components/navbar";

import { LeagueProvider } from './context/LeagueContext';
import { UserProvider } from './context/UserContext';
import Dashboard from './pages/Dashboard';
import Login from "./pages/Login";
import Draftboard from './pages/Draftboard';
import TeamSheets from './pages/Teamsheets';
import DraftHistory from './pages/DraftHistory';
import JoinLeague from './pages/JoinLeague';
import { DraftHistoryProvider } from './context/DraftHistoryContext';
import MyLeagues from './pages/MyLeagues';

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
                    <Route path='/my-leagues' element={<MyLeagues />} />
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
