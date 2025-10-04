import LeagueDropdown from "../components/league-data"

const backcolor = {
    backgroundColor: '#1E2A38'
}
export default function Dashboard() {
    return (
        <html style={backcolor}>
            <div className="container-dashboard">
                <h1>Welcome Back, Player.</h1>
                <span>Leagues that you take part in will show up here.</span>
                <a href="https://google.com">(click here to join new league)</a>
                <LeagueDropdown />
            </div>

        </html>
    )
}
