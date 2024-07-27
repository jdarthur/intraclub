import * as React from "react"
import {isMatchWon} from "./SetScores";
import {MatchupProps} from "./Matchup";
import {
    CheckCircleFilled,
    CloseCircleFilled,
} from "@ant-design/icons";
import {Space} from "antd";
import {WinLossIcon} from "./WinLossIcon";

type WonLostTopLineProps = {
    Matchups: MatchupProps[]
    Home?: boolean
}

export function WonLostTopLine({Matchups, Home}: WonLostTopLineProps) {

    let wonCount = 0
    let lostCount = 0

    for (let i = 0; i < Matchups?.length; i++) {
        const m = Matchups[i]
        const us = Home ? m.Result.Us : m.Result.Them
        const them = Home ? m.Result.Them : m.Result.Us

        if (isMatchWon(us, them)) {
            wonCount += 1
        } else if (isMatchWon(them, us)) {
            lostCount += 1
        }
    }

    return <Space>
        <div style={{display: "flex", alignItems: "center"}}>
            {wonCount}
            <span style={{width: "0.25em"}}/>
            <CheckCircleFilled style={{fontSize: "0.75em"}}/>
            {/*<WinLossIcon win={true} sizeEm={0.75}/>*/}
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
            {lostCount}
            <span style={{width: "0.25em"}}/>
            <CloseCircleFilled style={{fontSize: "0.75em"}}/>
            {/*<WinLossIcon win={false} sizeEm={0.75}/>*/}
        </div>

    </Space>


}