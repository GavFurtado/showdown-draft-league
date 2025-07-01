import NavBar from "../componenets/navbar"
import DraftCard from "../componenets/draftCards"

export default function draftboard(){
    console.log(DraftCard)
    return(
        <>
        <div className="h-full bg-[#BFC0C0]" flex flex-row>
            
            <NavBar page="Draftboard"/>
            
            <div className="flex flex-row">
            <div className="grid grid-cols-4 gap-4 m-4 p-6 pr-8  w-[70%]  overflow-scroll h-screen rounded-2xl">
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
                <DraftCard/>
            </div>
            
            <div className="bg-white shadow-md rounded-md overflow-hidden w-[25%] mx-auto mt-16 ml-2 h-[100%]">
                <div className="bg-gray-100 py-2 px-4">
                    <h2 className="text-l font-semibold text-gray-800">Your Team</h2>
                </div>
                <ul className="divide-y divide-gray-200">
                    <li className="flex items-center py-4 px-6">
                        
                        <img className="w-12 h-12 object-cover mr-4" src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/9.png" alt="User avatar"></img>
                        <div className="flex-1">
                            <h3 className="text-lg font-medium text-gray-800">BIG MAN BLASTOISE</h3>
                            {/* <p className="text-gray-600 text-base">1234 points</p> */}
                        </div>
                    </li>
                            
                            
                </ul>
            </div>
            </div>
        </div>
        </>
    )
}