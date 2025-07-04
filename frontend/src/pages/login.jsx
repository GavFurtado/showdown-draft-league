import loginJpg from '../assets/login-bg.jpg'
import discordPng from '../assets/discord-logo.png'

const backgroundStyle={
        backgroundImage:`url('${loginJpg}')`,
        backgroundSize: 'cover',           
        backgroundPosition: 'center',        
        backgroundRepeat: 'no-repeat',       
        minHeight: '100vh',
        display: 'flex',     
        flexDirection: 'column', 
        justifyContent: 'center', 
        alignItems: 'center',
}
const loginStyle={
        background:"white",
        borderRadius:"10px",
        minHeight:"30vh",
        minWidth:"25vw",
}
const imageStyle={
        minHeight:"10vh",
        minWidth:"5vw",
        paddingBottom:"0px",
        marginBottom:'0px'

}

export default function Login(){
    function submit(event){
        event.preventDefault();
        //fetch to backend oauth

    }
    return (
        <>
        <div style={backgroundStyle}>
            <div style={loginStyle} class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8">
                <div class="sm:mx-auto sm:w-full sm:max-w-sm">
                    <img style={imageStyle}class="mx-auto h-10 w-auto" src={discordPng} alt="Discord Logo" />
                    <h2 class="mt-10 text-center text-2xl/9 font-old tracking-tight text-gray-900">Login with Discord</h2>
                </div>

                <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
                    <form onSubmit={submit} class="space-y-6" action="#" method="POST">
                    <div>
                        <button type="submit" class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600">Login</button>
                    </div>
                    </form>
                </div>
            </div>
        </div>
        </>
    )
};