import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { discordCallback, TOKEN_KEY } from '../api/api';

export default function Callback() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const [tokenInput, setTokenInput] = useState('');
    const [status, setStatus] = useState<'exchanging' | 'manual' | 'error'>('manual');
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        if (code && state) {
            setStatus('exchanging');
            discordCallback(code)
                .then((res) => {
                    const token = res.data?.Token;
                    if (token) {
                        localStorage.setItem(TOKEN_KEY, token);
                        navigate('/my-leagues');
                    } else {
                        setStatus('manual');
                    }
                })
                .catch(() => {
                    setStatus('manual');
                    setError('Auto-exchange failed. Paste your JWT manually.');
                });
        }
    }, [searchParams, navigate]);

    const handleManualLogin = () => {
        if (!tokenInput.trim()) return;
        localStorage.setItem(TOKEN_KEY, tokenInput.trim());
        navigate('/my-leagues');
    };

    if (status === 'exchanging') {
        return (
            <div className="min-h-screen flex flex-col justify-center items-center bg-background-main">
                <p className="text-text-primary text-lg">Completing login...</p>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex flex-col justify-center items-center bg-background-main">
            <div className="bg-background-surface rounded-lg min-h-[30vh] min-w-[25vw] flex flex-col justify-center px-6 py-12 lg:px-8">
                <div className="sm:mx-auto sm:w-full sm:max-w-sm">
                    <h2 className="text-center text-2xl/9 font-old tracking-tight text-text-primary">
                        Paste Your JWT
                    </h2>
                    <p className="mt-2 text-center text-sm text-text-secondary">
                        Authenticate at the{' '}
                        <a
                            href="http://localhost:4200"
                            target="_blank"
                            rel="noreferrer"
                            className="text-accent-primary hover:text-accent-primary-hover underline"
                        >
                            new client
                        </a>
                        , then grab the JWT from localStorage key{' '}
                        <code className="bg-background-tertiary px-1 rounded">jwt_token</code>.
                    </p>
                </div>

                {error && (
                    <p className="mt-4 text-center text-sm text-error-500">{error}</p>
                )}

                <div className="mt-6 sm:mx-auto sm:w-full sm:max-w-sm">
                    <textarea
                        value={tokenInput}
                        onChange={(e) => setTokenInput(e.target.value)}
                        placeholder="eyJhbGciOi..."
                        rows={4}
                        className="block w-full rounded-md bg-background-tertiary px-3 py-1.5 text-sm text-text-primary placeholder:text-text-secondary outline-1 outline-background-secondary focus:outline-2 focus:outline-accent-primary"
                    />
                    <button
                        onClick={handleManualLogin}
                        className="mt-4 flex w-full justify-center rounded-md bg-accent-primary px-3 py-1.5 text-sm/6 font-semibold text-text-on-accent shadow-xs hover:bg-accent-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent-primary"
                    >
                        Login
                    </button>
                </div>
            </div>
        </div>
    );
}
