import React from 'react';

interface LayoutProps {
    children: React.ReactNode;
    variant?: 'container' | 'full';
}

export default function Layout({ children, variant = 'container' }: LayoutProps) {
    return (
        <main className="min-h-screen bg-background-main">
            {variant === 'full' ? (
                children
            ) : (
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-8">
                    {children}
                </div>
            )}
        </main>
    );
}
