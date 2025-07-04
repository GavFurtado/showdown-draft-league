export default function filter(){
    return(
        <div className=" flex items-center gap-4 sm:gap-6">
            <details className="group relative">
                <summary
                className="flex items-center gap-2 border-b border-gray-300 pb-1 text-gray-700 transition-colors hover:border-gray-400 hover:text-gray-900 dark:border-gray-600 dark:text-gray-200 dark:hover:border-gray-700 dark:hover:text-white [&::-webkit-details-marker]:hidden"
                >
                <span className="text-sm text-black font-medium"> Type </span>

                <span className="transition-transform group-open:-rotate-180">
                    <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    strokeWidth="1.5"
                    stroke="black"
                    className="size-4"
                    >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                    </svg>
                </span>
                </summary>

                <div
                className="z-auto w-64 divide-y divide-gray-300 rounded border border-gray-300 bg-white shadow-sm group-open:absolute group-open:start-0 group-open:top-8 dark:divide-gray-600 dark:border-gray-600 dark:bg-gray-900"
                >
                <fieldset className="p-3">
                    <legend className="sr-only">Checkboxes</legend>

                    <div className="flex flex-col items-start gap-3">
                    <label htmlFor="Option1" className="inline-flex items-center gap-3">
                        <input
                        type="checkbox"
                        className="size-5 rounded border-gray-300 shadow-sm dark:border-gray-600 dark:bg-gray-900 dark:ring-offset-gray-900 dark:checked:bg-blue-600"
                        id="Option1"
                        />

                        <span className="text-sm font-medium text-gray-700 dark:text-gray-200"> Option 1 </span>
                    </label>

                    <label htmlFor="Option2" className="inline-flex items-center gap-3">
                        <input
                        type="checkbox"
                        className="size-5 rounded border-gray-300 shadow-sm dark:border-gray-600 dark:bg-gray-900 dark:ring-offset-gray-900 dark:checked:bg-blue-600"
                        id="Option2"
                        />

                        <span className="text-sm font-medium text-gray-700 dark:text-gray-200"> Option 2 </span>
                    </label>

                    <label htmlFor="Option3" className="inline-flex items-center gap-3">
                        <input
                        type="checkbox"
                        className="size-5 rounded border-gray-300 shadow-sm dark:border-gray-600 dark:bg-gray-900 dark:ring-offset-gray-900 dark:checked:bg-blue-600"
                        id="Option3"
                        />

                        <span className="text-sm font-medium text-gray-700 dark:text-gray-200"> Option 3 </span>
                    </label>
                    </div>
                </fieldset>
                </div>
            </details>

            <details className="group relative">
                <summary
                className="flex items-center gap-2 border-b border-gray-300 pb-1 text-gray-700 transition-colors hover:border-gray-400 hover:text-gray-900 dark:border-gray-600 dark:text-gray-200 dark:hover:border-gray-700 dark:hover:text-white [&::-webkit-details-marker]:hidden"
                >
                <span className="text-sm text-black font-medium">Cost</span>

                <span className="transition-transform group-open:-rotate-180">
                    <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    strokeWidth="1.5"
                    stroke="black"
                    className="size-4"
                    >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                    </svg>
                </span>
                </summary>

                <div
                className="z-auto w-64 divide-y divide-gray-300 rounded border border-gray-300 bg-white shadow-sm group-open:absolute group-open:start-0 group-open:top-8 dark:divide-gray-600 dark:border-gray-600 dark:bg-gray-900"
                >
                <div className="flex items-center justify-between px-3 py-2">
                    <span className="text-sm text-gray-700 dark:text-gray-200"> Max price is $600 </span>

                    <button
                    type="button"
                    className="text-sm text-gray-700 underline transition-colors hover:text-gray-900 dark:text-gray-200 dark:hover:text-white"
                    >
                    Reset
                    </button>
                </div>

                <div className="flex items-center gap-3 p-3">
                    <label htmlFor="MinPrice">
                    <span className="text-sm text-gray-700 dark:text-gray-200"> Min </span>

                    <input
                        type="number"
                        id="MinPrice"
                        value="0"
                        className="mt-0.5 w-full rounded border-gray-300 shadow-sm sm:text-sm dark:border-gray-600 dark:bg-gray-900 dark:text-white"
                    />
                    </label>

                    <label htmlFor="MaxPrice">
                    <span className="text-sm text-gray-700 dark:text-gray-200"> Max </span>

                    <input
                        type="number"
                        id="MaxPrice"
                        value="600"
                        className="mt-0.5 w-full rounded border-gray-300 shadow-sm sm:text-sm dark:border-gray-600 dark:bg-gray-900 dark:text-white"
                    />
                    </label>
                </div>
                </div>
            </details>
            <details className="group relative">
                <summary
                className="flex items-center gap-2 border-b border-gray-300 pb-1 text-gray-700 transition-colors hover:border-gray-400 hover:text-gray-900 dark:border-gray-600 dark:text-gray-200 dark:hover:border-gray-700 dark:hover:text-white [&::-webkit-details-marker]:hidden"
                >
                <span className="text-sm text-black font-medium">Stats</span>

                <span className="transition-transform group-open:-rotate-180">
                    <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    strokeWidth="1.5"
                    stroke="black"
                    className="size-4"
                    >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                    </svg>
                </span>
                </summary>

                <div
                className="z-auto w-64 divide-y divide-gray-300 rounded border border-gray-300 bg-white shadow-sm group-open:absolute group-open:start-0 group-open:top-8 dark:divide-gray-600 dark:border-gray-600 dark:bg-gray-900"
                >
                <div className="flex items-center justify-between px-3 py-2">
                    <span className="text-sm text-gray-700 dark:text-gray-200"> Max price is $600 </span>

                    <button
                    type="button"
                    className="text-sm text-gray-700 underline transition-colors hover:text-gray-900 dark:text-gray-200 dark:hover:text-white"
                    >
                    Reset
                    </button>
                </div>

                <div className="flex items-center gap-3 p-3">
                    <label htmlFor="MinPrice">
                    <span className="text-sm text-gray-700 dark:text-gray-200"> Min </span>

                    <input
                        type="number"
                        id="MinPrice"
                        value="0"
                        className="mt-0.5 w-full rounded border-gray-300 shadow-sm sm:text-sm dark:border-gray-600 dark:bg-gray-900 dark:text-white"
                    />
                    </label>

                    <label htmlFor="MaxPrice">
                    <span className="text-sm text-gray-700 dark:text-gray-200"> Max </span>

                    <input
                        type="number"
                        id="MaxPrice"
                        value="600"
                        className="mt-0.5 w-full rounded border-gray-300 shadow-sm sm:text-sm dark:border-gray-600 dark:bg-gray-900 dark:text-white"
                    />
                    </label>
                </div>
                </div>
            </details>
            
        </div>

    )
}