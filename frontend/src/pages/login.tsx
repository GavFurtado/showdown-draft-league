import loginJpg from '../assets/login-bg.jpg'
import discordPng from '../assets/discord-logo.png'
import { getDiscordLoginUrl } from '../api/api'

export default function Login() {
    const handleLogin = () => {
        window.location.href = getDiscordLoginUrl();
    };
    return (
        <>
            <div
                style={{ backgroundImage: `url('${loginJpg}')` }}
                className="bg-cover bg-center bg-no-repeat min-h-screen flex flex-col justify-center items-center"
            >
                <div className="bg-background-surface rounded-lg min-h-[30vh] min-w-[25vw] flex flex-col justify-center px-6 py-12 lg:px-8">
                    <div className="sm:mx-auto sm:w-full sm:max-w-sm">
                        <img className="min-h-[10vh] min-w-[5vw] pb-0 mb-0 mx-auto h-10 w-auto" src={discordPng} alt="Discord Logo" />
                        <h2 className="mt-10 text-center text-2xl/9 font-old tracking-tight text-text-primary">Login with Discord</h2>
                    </div>

                    <div className="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
                        <div>
                            <button onClick={handleLogin} className="flex w-full justify-center rounded-md bg-accent-primary px-3 py-1.5 text-sm/6 font-semibold text-text-on-accent shadow-xs hover:bg-accent-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-primary">Login</button>
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
};
