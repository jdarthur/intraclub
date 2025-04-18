import {MatchupProps} from "./Matchup";
import {Card, ColorPicker, Input, Popover} from "antd";
import * as React from "react";
import {calculateScores, CARD_GAP_EM, CARD_WIDTH} from "./Scoreboard";
import {useUpdateTeamInfoMutation} from "../redux/api";
import {useSearchParams} from "react-router-dom";
import {stringEditorDisplayType} from "./Player";
import {CloseOutlined, HomeOutlined, TruckOutlined} from "@ant-design/icons";
import {WonLostTopLine} from "./WonLostTopLine";
import {useEffect, useState} from "react";
import {TooltipFocus} from "./TooltipFocus";

export type Team = {
    name: string
    color: string
}

function TeamColorDisplay({value, setValue, onSave, readOnly}: stringEditorDisplayType) {

    // on change, set the color value on the parent component to the hex of the selected color in the picker
    const onChange = (v: any, hex: string) => {
        setValue(hex)
    }

    const [open, setOpen] = useState<boolean>(false)

    const onClose = () => {
        console.log("set color: ", value)
        setOpen(false)
        onSave()
    }

    return <div style={{display: "flex", alignItems: "center"}}>
        <TooltipFocus open={open} zIndex={3} onClose={onClose}/>

        <div onClick={() => setOpen(true)} style={{display: "flex", alignItems: "center"}}>
            <ColorPicker value={value} onChange={onChange} disabled={readOnly} open={open}
                         size={"large"}/>
        </div>
    </div>
}

function TeamNameDisplay({value, setValue, onSave, readOnly}: stringEditorDisplayType) {

    // on change, set the name value on the parent component to the provided name
    const onChange = (event: any) => {
        setValue(event.target.value)
    }

    const [open, setOpen] = useState<boolean>(false)

    const onClose = () => {
        console.log("Update team name: ", value)
        setOpen(false)
        onSave()
    }


    // name value will show a grey placeholder value when unset
    const nameValue = value ? value : <span style={{color: "#bfbfbf"}}> Name not set </span>

    // this is what's primarily displayed in the modal
    const name = <span
        style={{minWidth: 100, minHeight: "50px", cursor: readOnly ? "auto" : "pointer"}}
        onClick={() => setOpen(true)}>
        {nameValue}
    </span>

    // don't display a popover when we are in read-only mode
    if (readOnly) {
        return name
    }

    const title = <span
        style={{fontSize: "1.4em", display: "flex", alignItems: "center", justifyContent: "space-between"}}>
        Set team name
        <CloseOutlined onClick={onClose} style={{color: "rgba(0, 0, 0, 0.4)"}}/>
    </span>

    // content inside of the popover
    const content = <Input value={value}
                           style={{fontSize: "1.4em"}}
                           onChange={onChange}
                           ref={el => {
                               setTimeout(() => el?.focus(), 0); // autofocus the input
                           }}/>


    return <div>
        <TooltipFocus open={open} zIndex={3} onClose={onClose}/>
        <Popover title={title} content={content} open={open} trigger={"click"} zIndex={4}>
            {name}
        </Popover>
    </div>

}

type OneTeamScoreProps = {
    Matchups: MatchupProps[]
    Team: Team,
    Home: boolean
    NarrowScreen?: boolean
}

export function OneTeamScore({Matchups, Team, Home, NarrowScreen}: OneTeamScoreProps) {
    const [team, setTeam] = React.useState<Team>(Team)
    const [updateTeam] = useUpdateTeamInfoMutation()

    useEffect(() => {
        const v: Team = {
            name: Team.name,
            color: Team.color,
        }
        setTeam(v)
    }, [Team])

    // get the `key` value from the query params which determines if we are in read-only mode
    const [searchParams] = useSearchParams()
    const key = searchParams.get('key')

    const setColor = (value: string) => {
        const t: Team = {...team}
        t.color = value
        setTeam(t)
    }

    const setName = (value: string) => {
        const t: Team = {...team}
        t.name = value
        setTeam(t)
    }

    const onSave = () => {
        const body = {
            home: Home,
            name: team.name,
            color: team.color,
            key: key
        }
        console.log(`save team (home=${Home})`, body)
        updateTeam(body)
    }


    return <Card style={{marginBottom: "0.5em", flexGrow: 1}}
                 styles={{body: {padding: "0.25em 0.5em"}}}>
        <div style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            fontSize: NarrowScreen ? "2.5em" : "3.5em",
            background: "white",
            padding: NarrowScreen ? "0em 0.1em" : "0em 0.25em"
        }}>
            <div style={{display: "flex", alignItems: "center"}}>
                <div style={{
                    color: "rgba(0, 0, 0, 0.5)",
                    fontSize: "0.5em",
                    marginRight: "0.5em",
                }}>
                    {Home ? <HomeOutlined style={{marginTop: "0.4em"}}/> :
                        <TruckOutlined style={{marginTop: "0.4em", fontSize: "1.2em"}}/>}
                </div>

                <div style={{marginRight: "0.5em", display: "flex", alignItems: "center"}}>
                    <TeamColorDisplay value={team.color} setValue={setColor} onSave={onSave} readOnly={!key}/>
                </div>

                <div>
                    <TeamNameDisplay value={team.name} setValue={setName} onSave={onSave} readOnly={!key}/>
                </div>

            </div>


            <div style={{fontWeight: "bold", display: "inline-flex", alignItems: "center"}}>
                <span
                    style={{
                        color: "rgba(0, 0, 0, 0.3)",
                        fontSize: "0.6em",
                        marginTop: "0.2em",
                        marginRight: NarrowScreen ? "0.7em" : "1em",
                    }}>
                    <WonLostTopLine Matchups={Matchups} Home={Home}/>
                </span>

                {calculateScores(Matchups, Home)}
            </div>
        </div>
    </Card>
}