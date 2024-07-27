import {Team} from "./OneTeamScore";
import * as React from "react";
import {Pairing, PairingProps} from "./Pairing"
import {MatchProps} from "./SetScores";
import {CARD_GAP_EM, CARD_WIDTH} from "./Scoreboard";

export type MatchupProps = {
    HomePairing: PairingProps
    AwayPairing: PairingProps
    HomeTeam?: Team
    AwayTeam?: Team
    Result?: MatchProps
    NarrowScreen?: boolean
    WindowWidth?: number
    WindowHeight?: number
}

export function Matchup({
                            HomePairing,
                            AwayPairing,
                            Result,
                            HomeTeam,
                            AwayTeam,
                            NarrowScreen,
                            WindowWidth,
                        }: MatchupProps) {

    const width = WindowWidth ? (`calc(${WindowWidth / 3}px - ${CARD_GAP_EM + 1}em)`) : CARD_WIDTH

    return <div style={{
        display: "flex",
        alignItems: "center",
        flexWrap: "wrap",
        flexDirection: "column",
        width: NarrowScreen ? "100%" : width,
        margin: NarrowScreen ? "1em" : "0em"
    }}>
        <Pairing Result={Result}
                 Home={true}
                 Color={HomeTeam.color}
                 NarrowScreen={NarrowScreen}
                 player1={HomePairing.player1}
                 player2={HomePairing.player2}
        />
        <span style={{height: "0.5em"}}/>
        <Pairing Result={Result}
                 Home={false}
                 Color={AwayTeam.color}
                 NarrowScreen={NarrowScreen}
                 player1={AwayPairing.player1}
                 player2={AwayPairing.player2}
        />

    </div>
}




