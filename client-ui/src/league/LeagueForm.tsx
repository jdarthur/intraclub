import * as React from 'react';
import {TeamColorProps} from "../team/TeamColor";
import {
    useCreateLeagueMutation,
    useGetFacilitiesQuery,
    useUpdateLeagueMutation,
    useWhoAmIQuery
} from "../redux/api.js";
import {CommonFormModal, SubmitResult} from "../common/CommonFormModal";
import {
    InputFormItem,
    SelectFormItem,
    TimePickerFormItem
} from "../common/FormItem";
import {Facility} from "../model/facility";
import dayjs from "dayjs";
import {useState} from "react";
import {WeekSelect, WeekSelectState} from "./WeekSelect";
import {Form} from "antd";
import {ColorSelect, TeamColor} from "./ColorSelect";
import {StepForm, StepFormStep} from "../common/StepForm";
import {StepFormModal} from "../common/StepFormModal";

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

export function LeagueForm({Update, InitialState, LeagueId}: LeagueFormProps) {

    const [weekIds, setWeekIds] = useState<string[]>(InitialState?.weeks || [])
    const [colors, setColors] = useState<TeamColor[]>(InitialState?.colors || [])
    const [disabled, setDisabled] = useState<boolean>(false)

    //const [formIsOpen, setFormIsOpen] = useState()
    const [form] = Form.useForm()

    const {data} = useWhoAmIQuery()

    const [createLeague] = useCreateLeagueMutation()
    const [updateLeague] = useUpdateLeagueMutation()
    const {data: facilities} = useGetFacilitiesQuery()

    const newInitialState = transformState(InitialState)

    const facilityOptions = facilities?.resource?.map((facility: Facility) => ({
        label: facility.name,
        value: facility.id
    }))

    const onSave = async (): Promise<SubmitResult> => {

        const formValues = form.getFieldsValue(true)
        console.log(formValues)

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

        return func()
    }

    const mainInfo = <div>
        <InputFormItem name={"name"} label={"League name"} disabled={disabled}/>
        <SelectFormItem name={"facility"} label={"Facility"} options={facilityOptions} disabled={disabled}/>
        <TimePickerFormItem name={"start_time"} label={"Start time"} disabled={disabled}/>
    </div>

    const weekSelect = <WeekSelect originalWeekIds={weekIds} setWeekIds={setWeekIds}
                                   update={Update} disabled={disabled}/>

    const colorSelect = <ColorSelect colors={colors} setColors={setColors} disabled={disabled}/>

    const review = <div>
        {mainInfo}
        {weekSelect}
        {colorSelect}
    </div>


    const steps: StepFormStep[] = [
        {title: "Basic info", content: mainInfo},
        {title: "Weeks", content: weekSelect},
        {title: "Colors", content: colorSelect},
        {title: "Review", content: review}
    ]

    return <StepFormModal ObjectType={"league"} IsUpdate={Update} InitialState={newInitialState}
                          form={form} footer={null} onStepFormFinish={onSave} steps={steps}
                          setDisabled={setDisabled} children={null} onCancel={onSave}/>
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

