import StandingsTable from "../componenets/standings-table";
import Score from "../componenets/score";
import Roster from "../componenets/roster";
import PlayerScedhule from "../componenets/playerScedhule";

export default function Dashboard(){
  return (
  <div className="container-dashboard">
      <div className="player-welcome"><h1>Welcome Back, Player.</h1></div>
      <div className="table-1"> 
        {/* table-1 for standings */}
        <StandingsTable />
      </div>
      <div className="table-2 ">
        {/* table 2 for scores */}
        <Score />
      </div>
      <div className="table-3">
        {/* table 3 for sced */}
        <PlayerScedhule />
      </div>
      <div className="table-4">
        {/* table 4 for roster */}  
          <Roster />
      </div>
    </div>
)}