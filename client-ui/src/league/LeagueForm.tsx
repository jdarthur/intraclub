import * as React from 'react';
import {TeamColorProps} from "../team/TeamColor";
import {
    useCreateLeagueMutation,
    useGetFacilitiesQuery,
    useUpdateLeagueMutation,
    useWhoAmIQuery
} from "../redux/api.js";
import {CommonFormModal} from "../common/CommonFormModal";
import {
    InputFormItem,
    SelectFormItem,
    TimePickerFormItem
} from "../common/FormItem";
import {Facility} from "../settings/Facilities";
import dayjs from "dayjs";
import {useState} from "react";
import {WeekSelect} from "./WeekSelect";
import {Button, Steps} from "antd";
import {ColorSelect, TeamColor} from "./ColorSelect";

export type League = {
    league_id?: string
    name: string
    colors?: TeamColorProps[]
    commissioner: string
    reporters?: string[]
    facility: string
    start_time: number
    weeks: string[]
    active?: boolean
}

type LeagueFormProps = {
    Update?: boolean
    InitialState?: League
    LeagueId?: string
}

const MAIN_INFO = 0
const WEEKS = 1
const COLORS = 2
const REVIEW = 3

export function LeagueForm({Update, InitialState, LeagueId}: LeagueFormProps) {

    const [weekIds, setWeekIds] = useState<string[]>(InitialState?.weeks || [])
    const [colors, setColors] = useState<TeamColor[]>(InitialState?.colors || [])

    const [step, setStep] = useState<number>(0);

    const disabled = step == REVIEW

    const {data} = useWhoAmIQuery()

    const [createLeague] = useCreateLeagueMutation()
    const [updateLeague] = useUpdateLeagueMutation()
    const {data: facilities} = useGetFacilitiesQuery()

    const newInitialState = transformState(InitialState)

    const facilityOptions = facilities?.resource?.map((facility: Facility) => ({
        label: facility.name,
        value: facility.id
    }))

    const onSave = async (formValues: any) => {
        const body: League = {
            name: formValues.name,
            commissioner: data?.user_id,
            weeks: weekIds,
            start_time: formValues.start_time?.format("HH:mm"),
            facility: formValues.facility,
            colors: colors,
        }

        let func = () => createLeague(body)
        if (Update) {
            func = () => updateLeague({id: LeagueId, body: body})
        }

        return await func()
    }

    const mainInfo = <div>
        <InputFormItem name={"name"} label={"League name"} disabled={disabled}/>
        <SelectFormItem name={"facility"} label={"Facility"} options={facilityOptions} disabled={disabled}/>
        <TimePickerFormItem name={"start_time"} label={"Start time"} disabled={disabled}/>
    </div>

    const weekSelect = <WeekSelect originalWeekIds={weekIds} setWeekIds={setWeekIds}
                                   leagueId={LeagueId} update={Update} disabled={disabled}/>

    const colorSelect = <ColorSelect colors={colors} setColors={setColors} disabled={disabled}/>

    const review = <div>
        {mainInfo}
        {weekSelect}
        {colorSelect}
    </div>


    const steps = [
        {title: "Basic information"},
        {title: "Weeks"},
        {title: "Colors"},
        {title: "Review"}
    ]

    const next = () => {

    }

    return <CommonFormModal ObjectType={"league"} OnSubmit={onSave} IsUpdate={Update} InitialState={newInitialState}>
        <Steps items={steps} current={step}/>
        <div style={{marginBottom: "1.5em"}}/>

        {step == MAIN_INFO ? mainInfo : null}
        {step == WEEKS ? weekSelect : null}
        {step == COLORS ? colorSelect : null}
        {step == REVIEW ? review : null}

        <Button onClick={() => setStep(step - 1)} disabled={step == MAIN_INFO}>
            Back
        </Button>
        <Button onClick={() => setStep(step + 1)} disabled={step == REVIEW}>
            Next
        </Button>

    </CommonFormModal>
}

type RealFormState = {
    name: string
    facility: string
    start_time: dayjs.Dayjs
}

function transformState(original: League): RealFormState {
    return {
        name: original?.name,
        facility: original?.facility,
        start_time: original?.start_time ? dayjs(original?.start_time, "HH:mm") : dayjs("08:30", "HH:mm"),
    }
}

