import {Team} from "./TeamName";
import * as React from "react";
import {Pairing, PairingProps} from "./Pairing"
import {MatchProps} from "./SetScores";
import {CARD_WIDTH} from "./Scoreboard";

export type MatchupProps = {
    HomePairing: PairingProps
    AwayPairing: PairingProps
    HomeTeam?: Team
    AwayTeam?: Team
    Result?: MatchProps
    NarrowScreen?: boolean
}

export function Matchup({HomePairing, AwayPairing, Result, HomeTeam, AwayTeam, NarrowScreen}: MatchupProps) {
    return <div style={{
        display: "flex",
        alignItems: "center",
        flexWrap: "wrap",
        flexDirection: "column",
        width: NarrowScreen ? "100%" : CARD_WIDTH,
    }}>
        <Pairing Result={Result}
                 Home={true}
                 Color={HomeTeam.color}
                 NarrowScreen={NarrowScreen}
                 player1={HomePairing.player1}
                 player2={HomePairing.player2}
        />
        <span style={{fontSize: "0.8em", color: "rgba(0,0,0,0.6)"}}>
            vs.
        </span>
        <Pairing Result={Result}
                 Home={false}
                 Color={AwayTeam.color}
                 NarrowScreen={NarrowScreen}
                 player1={AwayPairing.player1}
                 player2={AwayPairing.player2}
        />

    </div>
}




