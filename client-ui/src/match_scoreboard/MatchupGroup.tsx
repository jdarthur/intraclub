import * as React from "react";
import {Matchup, MatchupProps} from "./Matchup";

type MatchupGroupProps = {
    Matchups: MatchupProps[]
    NarrowScreen: boolean,
    ScreenWidth: number
}

export function MatchupGroup({Matchups, NarrowScreen, ScreenWidth}: MatchupGroupProps) {

    const matchups = Matchups?.map((m, i) => {
            return <Matchup HomePairing={m.HomePairing}
                            AwayPairing={m.AwayPairing}
                            HomeTeam={m.HomeTeam}
                            AwayTeam={m.AwayTeam}
                            Result={m.Result}
                            NarrowScreen={NarrowScreen}
                            key={`matchup${i}`}
                            WindowWidth={ScreenWidth}
            />
        }
    )


    return <div
        style={{
            display: "flex",
            alignItems: "stretch",
            flexWrap: "wrap",
            justifyContent: "space-evenly",
            height: "calc(50% - 2em)",
            width: "100%"
        }}>
        {matchups}
    </div>


}