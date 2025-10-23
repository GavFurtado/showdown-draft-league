import { useState } from "react";
import { DraftedPokemon } from "../api/data_interfaces"
import { formatPokemonName } from "../utils/nameFormatter";

interface SpeedTableProps {
    roster: DraftedPokemon[];
}

const SpeedTable: React.FC<SpeedTableProps> = ({ roster }) => {
    const [iVs, setIVs] = useState<number>(0);
    const [level, setLevel] = useState<number>(100);

    const handleChangeIVs = (e: any) => {
        setIVs(parseInt(e.target.value));
    }
    const handleChangeLevel = (e: any) => {
        let value = parseInt(e.target.value);
        if (value >= 45 && value <= 55) {
            value = 50;
        }
        setLevel(value);
    }

    const calcSpeedStat = (
        BaseSpeed: number,
        IVs: number,
        Level: number,
        EVs: number,
        Nature: "NEUTRAL" | "POSITIVE",
        Multiplier: 1.0 | 1.5 | 2.0
    ): number => {
        const evPart = Math.floor(EVs / 4);
        // Core speed formula
        let speed = Math.floor(((2 * BaseSpeed + IVs + evPart) * Level) / 100) + 5;
        // Apply nature (1.1 for positive)
        if (Nature === "POSITIVE") {
            speed = Math.floor(speed * 1.1);
        }
        // Apply Choice Scarf or Tailwind multiplier
        speed = Math.floor(speed * Multiplier);

        return speed;
    };

    return (
        <div className="w-full bg-background-surface p-4 rounded-lg shadow-md">
            <div className="flex justify-between items-center mb-4">
                <h2 className="items-left text-2xl font-bold text-text-primary">Speed Benchmarks</h2>
                <div className="flex items-center space-x-4">
                    <div className="flex items-center space-x-2">
                        <p className="text-text-primary font-bold align-middle text-sm">IVs: {iVs}</p>
                        <input
                            type="range"
                            min="0"
                            max="31"
                            step="1"
                            value={iVs}
                            onChange={handleChangeIVs}
                            className="relative z-10 w-24 h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer"
                        />
                    </div>
                    <div className="flex items-center space-x-2">
                        <p className="text-text-primary font-bold align-middle text-sm">Level: {level}</p>
                        <input
                            type="range"
                            min="1"
                            max="100"
                            step="1"
                            value={level}
                            onChange={handleChangeLevel}
                            className="relative z-10 w-24 h-2 bg-slate-200 rounded-lg appearance-none cursor-pointer"
                        />
                    </div>
                </div>
            </div>
            <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-background-primary">
                        <tr>
                            <th scope="col" className="px-2 py-4 w-[165px] text-text-on-nav text-left text-[10px] font-bold uppercase tracking-wider w-60">Pokémon ({roster.length})</th>
                            <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold tracking-wider">Base Speed</th>
                            <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold tracking-wider">0 EVs</th>
                            <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold tracking-wider">252 EVs</th>
                            <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold tracking-wider">252 EVs|+Spe</th>
                            <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold tracking-wider">252 EVs|+Spe|Scarf(+1)</th>
                            <th scope="col" className="px-2 py-4 text-text-on-nav text-left text-[10px] font-bold tracking-wider">252 EVs|+Spe|Tailwind(+2)</th>
                        </tr>
                    </thead>
                    <tbody className="bg-white">
                        {roster.map(dp => (
                            <tr key={dp.ID} className='border-b border-gray-900/10'>
                                <td className="px-2 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                                    <div className="flex items-center">
                                        <img src={dp.PokemonSpecies.Sprites.FrontDefault} alt={dp.PokemonSpecies.Name} className="h-8 w-8 mr-2" />
                                        <span className={dp.PokemonSpecies.Name.length > 14 ? 'text-xs' : 'text-sm'}>
                                            {formatPokemonName(dp.PokemonSpecies.Name)}
                                        </span>
                                    </div>
                                </td>
                                <td
                                    className="px-2 py-3 whitespace-nowrap text-left text-sm font-semibold text-gray-500"
                                >
                                    {dp.PokemonSpecies.Stats.Speed}
                                </td>
                                <td
                                    className="px-2 py-3 whitespace-nowrap text-left text-sm font-semibold text-gray-500"
                                >
                                    {calcSpeedStat(dp.PokemonSpecies.Stats.Speed, iVs, level, 0, "NEUTRAL", 1)}
                                </td>
                                <td
                                    className="px-2 py-3 whitespace-nowrap text-left text-sm font-semibold text-gray-500"
                                >
                                    {calcSpeedStat(dp.PokemonSpecies.Stats.Speed, iVs, level, 252, "NEUTRAL", 1)}
                                </td>
                                <td
                                    className="px-2 py-3 whitespace-nowrap text-left text-sm font-semibold text-gray-500"
                                >
                                    {calcSpeedStat(dp.PokemonSpecies.Stats.Speed, iVs, level, 252, "POSITIVE", 1)}
                                </td>
                                <td
                                    className="px-2 py-3 whitespace-nowrap text-left text-sm font-semibold text-gray-500"
                                >
                                    {calcSpeedStat(dp.PokemonSpecies.Stats.Speed, iVs, level, 252, "POSITIVE", 1.5)}
                                </td>
                                <td
                                    className="px-2 py-3 whitespace-nowrap text-left text-sm font-semibold text-gray-500"
                                >
                                    {calcSpeedStat(dp.PokemonSpecies.Stats.Speed, iVs, level, 252, "POSITIVE", 2.0)}
                                </td>
                            </tr>
                        ))
                        }
                    </tbody>
                </table>
            </div>
        </div>
    )
}


export default SpeedTable;
