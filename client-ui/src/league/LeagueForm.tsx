import * as React from 'react';
import {Form} from "antd";
import {TeamColorProps} from "../team/TeamColor";
import {useCreateLeagueMutation, useCreateWeekMutation, useGetFacilitiesQuery, useWhoAmIQuery} from "../redux/api.js";
import {CommonModal} from "../common/CommonModal";
import {
    DatePickerFormItem,
    InputFormItem,
    SelectFormItem,
    TimePickerFormItem
} from "../common/FormItem";
import {Facility} from "../settings/Facilities";

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

type Week = {
    week_id?: string
    date: string
    original_date: string
}

type LeagueFormProps = {
    Update?: boolean
    InitialState?: League
}

export function LeagueForm() {

    const [form] = Form.useForm()

    const {data} = useWhoAmIQuery()

    const [createWeek] = useCreateWeekMutation()
    const [createLeague] = useCreateLeagueMutation()
    const {data: facilities} = useGetFacilitiesQuery()

    const facilityOptions = facilities?.resource?.map((facility: Facility) => ({
        label: facility.name,
        value: facility.id
    }))


    const onSave = async () => {
        const formValues = form.getFieldsValue();

        const createWeeks = await CreateWeek(formValues.weeks, createWeek)
        if (!createWeeks.success) {
            return {
                error: createWeeks.error
            }
        }

        const body: League = {
            name: formValues.name,
            commissioner: data?.user_id,
            weeks: createWeeks.weekIds,
            start_time: formValues.start_time?.format("HH:mm"),
            facility: formValues.facility,
        }

        return await createLeague(body)
    }

    return <CommonModal ObjectType={"league"} OnSubmit={onSave} IsUpdate={false}>
        <Form form={form}>
            <InputFormItem name={"name"} label={"Name"}/>
            <SelectFormItem name={"facility"} label={"Facility"} options={facilityOptions}/>
            <TimePickerFormItem name={"start_time"} label={"Start time"}/>
            <DatePickerFormItem name={"weeks"} label={"Weeks"} future multiple/>
        </Form>
    </CommonModal>
}

type CreateWeeksResult = {
    success: boolean
    error?: string
    weekIds?: string[]
}

async function CreateWeek(weeks: any[], createWeek: (v: any) => Promise<any>): Promise<CreateWeeksResult> {

    const weekIds: string[] = []

    for (let i = 0; i < weeks.length; i++) {
        // format the dayjs value as a YYYY-MM-DD date
        const date = weeks[i]?.format("YYYY-MM-DD")

        // create a POST body for a new Week record
        const body: Week = {
            date: date,
            original_date: date
        }

        const res = await createWeek(body)
        if (res.error) {
            return {
                success: false,
                error: res.error
            }
        }

        weekIds.push(res?.data?.resource?.week_id)
    }

    return {
        success: true,
        weekIds: weekIds,
    }
}

