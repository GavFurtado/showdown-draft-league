import {useEffect, useRef } from 'react';

const pokemonTypes = [
    'normal', 'fire', 'water', 'grass', 'electric', 'ice', 'fighting',
    'poison', 'ground', 'flying', 'psychic', 'bug', 'rock', 'ghost',
    'dragon', 'steel', 'dark', 'fairy',
];

const pokemonStats = [
    { key: 'HP', name: 'hp' },
    { key: 'Attack', name: 'attack' },
    { key: 'Defense', name: 'defense' },
    { key: 'Sp. Attack', name: 'special-attack' },
    { key: 'Sp. Defense', name: 'specaial-defence' },
    { key: 'Speed', name: 'speed' },
];

export default function Filter(props) {
    const {selectedTypes, selectedCost, sortByStat, sortOrder} = props.filters
    const updateFilter = props.updateFilter
    const resetAllFilters = props.resetAllFilters

    function handleTypeChange(type){
        let newSelectedTypes
        if (selectedTypes.includes(type)) {
            newSelectedTypes = selectedTypes.filter(t => t !== type);
        } else {
            newSelectedTypes = [...selectedTypes, type];
        }
        updateFilter("selectedTypes",newSelectedTypes)
    }

    const handleSelectedCostChange = (event) => {
        const value = event.target.value;
        if (value === '') {
            updateFilter("selectedCost","")
        } else {
            const numValue = parseInt(value, 10);
            if (!isNaN(numValue) && numValue >= 1 && numValue <= 20) {
                updateFilter("selectedCost",numValue);
            } else if (numValue < 1) {
                updateFilter("selectedCost",1);
            } else if (numValue > 20) {
                updateFilter("selectedCost",20);
            } else {
                updateFilter("selectedCost","");
            }
        }
    };

    const handleSortByStatChange = (event) => {
        const statObj = pokemonStats.find(x => x.key === event.target.value);
        const statName = statObj ? statObj.name : null;
        updateFilter("sortByStat", statName); 
    };
    const handleSortOrderChange = (event) => {
        updateFilter("sortOrder",event.target.value);
    };

   
    const typeRef = useRef();
    const costRef = useRef();
    const statRef = useRef();
    useEffect(() => {
        function handleClickOutside(event) {
            if (typeRef.current && !typeRef.current.contains(event.target)) {
                typeRef.current.removeAttribute('open');
            }
            if (costRef.current && !costRef.current.contains(event.target)) {
                costRef.current.removeAttribute('open');
            }
            if (statRef.current && !statRef.current.contains(event.target)) {
                statRef.current.removeAttribute('open');
            }
        }

        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);




    return (
        <div className="flex items-center gap-4 sm:gap-6">
            <details className="group relative" ref={typeRef}>
                <summary
                    className={`flex items-center gap-2 pb-1 text-[#2D3142] [&::-webkit-details-marker]:hidden
                                border-b-2 border-transparent`}
                >
                    <span className="text-sm font-medium"> Type ({selectedTypes.length}) </span>

                    <span className="transition-transform group-open:-rotate-180">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth="1.5"
                            stroke="#2D3142"
                            className="size-4"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                        </svg>
                    </span>
                </summary>

                <div
                    className={`z-50 w-64 divide-y divide-[#4F5D75] rounded border border-[#4F5D75]
                               bg-[#2D3142] shadow-sm group-open:absolute group-open:start-0 group-open:top-8`}
                >
                    <fieldset className="p-3">
                        <legend className="sr-only">Pokemon Types</legend>

                        <div className="grid grid-cols-3 gap-3">
                            {pokemonTypes.map((type) => (
                                <label
                                    key={type}
                                    htmlFor={type}
                                    className={`inline-flex items-center gap-3 cursor-pointer justify-center
                                            rounded-md p-1 text-white
                                            hover:bg-[#BFC0C0]
                                            has-[input:checked]:bg-[#EF8354] has-[input:checked]:text-[#2D3142]`}
                                >
                                    <input
                                        type="checkbox"
                                        className="size-5 rounded border-gray-300 shadow-sm sr-only"
                                        id={type}
                                        checked={selectedTypes.includes(type)}
                                        onChange={() => handleTypeChange(type)}
                                    />
                                    <span className="text-sm font-medium"> {type.charAt(0).toUpperCase() + type.slice(1)} </span>
                                </label>
                            ))}
                        </div>
                    </fieldset>
                </div>
            </details>

            <details className="group relative" ref={costRef}>
                <summary
                    className={`flex items-center gap-2 pb-1 text-[#2D3142] [&::-webkit-details-marker]:hidden
                                border-b-2 border-transparent`}
                >
                    <span className="text-sm font-medium">Cost {selectedCost ? `(${selectedCost})` : ''}</span>

                    <span className="transition-transform group-open:-rotate-180">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth="1.5"
                            stroke="#2D3142"
                            className="size-4"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                        </svg>
                    </span>
                </summary>

                <div
                    className={`z-50 w-64 divide-y divide-[#4F5D75] rounded border border-[#4F5D75]
                               bg-[#2D3142] shadow-sm group-open:absolute group-open:start-0 group-open:top-8`}
                >
                    <div className="flex flex-col p-3">
                        <label htmlFor="costInput" className="block text-sm font-medium text-white mb-1">Enter Cost (1-20):</label>
                        <input
                            type="number"
                            id="costInput"
                            min="1"
                            max="20"
                            value={selectedCost}
                            onChange={handleSelectedCostChange}
                            className={`w-full rounded border-[#4F5D75] shadow-sm sm:text-sm
                                        bg-[#4F5D75] text-white focus:ring-[#EF8354] focus:border-[#EF8354]`}
                            placeholder="e.g., 10"
                        />
                        <button
                            type="button"
                            onClick={() => updateFilter("selectedCost",'')}
                            className={`text-sm underline transition-colors text-[#BFC0C0] hover:text-white mt-3 self-end`}
                        >
                            Reset
                        </button>
                    </div>
                </div>
            </details>

            <details className="group relative" ref={statRef}>
                <summary
                    className={`flex items-center gap-2 pb-1 text-[#2D3142] [&::-webkit-details-marker]:hidden
                                border-b-2 border-transparent`}
                >
                    <span className="text-sm font-medium">Stats</span>

                    <span className="transition-transform group-open:-rotate-180">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth="1.5"
                            stroke="#2D3142"
                            className="size-4"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                        </svg>
                    </span>
                </summary>

                <div
                    className={`z-50 w-64 divide-y divide-[#4F5D75] rounded border border-[#4F5D75]
                               bg-[#2D3142] shadow-sm group-open:absolute group-open:start-0 group-open:top-8`}
                >
                    <div className="flex flex-col p-3 gap-3">
                        <div>
                            <label htmlFor="sortByStat" className="block text-sm font-medium text-white mb-1">Sort By:</label>
                            <select
                                id="sortByStat"
                                value={pokemonStats.find(x => x.name === sortByStat)?.key || ''}
                                onChange={handleSortByStatChange}
                                className={`mt-0.5 w-full rounded border-[#4F5D75] shadow-sm sm:text-sm
                                            bg-[#4F5D75] text-white focus:ring-[#EF8354] focus:border-[#EF8354]`}
                            >
                                <option value="">None</option>
                                {pokemonStats.map(stat => (
                                    <option key={stat.key} value={stat.key}>{stat.key}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label htmlFor="sortOrder" className="block text-sm font-medium text-white mb-1">Order:</label>
                            <select
                                id="sortOrder"
                                value={sortOrder}
                                onChange={handleSortOrderChange}
                                className={`mt-0.5 w-full rounded border-[#4F5D75] shadow-sm sm:text-sm
                                            bg-[#4F5D75] text-white focus:ring-[#EF8354] focus:border-[#EF8354]`}
                            >
                                <option value="asc">Ascending</option>
                                <option value="desc">Descending</option>
                            </select>
                        </div>
                    </div>
                    <div className="flex items-center justify-end px-3 py-2">
                        <button
                            type="button"
                            onClick={() => { updateFilter("sortByStat",''); updateFilter("sortOrder",'asc'); }}
                            className={`text-sm underline transition-colors text-[#BFC0C0] hover:text-white`}
                        >
                            Reset
                        </button>
                    </div>
                </div>
            </details>

            <button
                onClick={resetAllFilters}
                className={`py-1 px-3 rounded-md text-sm font-medium border  hover:bg-[#2D3142] hover:text-white transition-colors`}
            >
                Clear All Filters
            </button>
        </div>
    );
}