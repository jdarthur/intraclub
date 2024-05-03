import {useEffect, useState} from "react";
import {useCreateWeekMutation, useDeleteWeekMutation, useGetWeeksByLeagueIdQuery} from "../redux/api";
import dayjs from "dayjs";
import {FormItem} from "../common/FormItem";
import {Button, DatePicker} from "antd";
import * as React from "react";

export type Week = {
    week_id?: string
    date: string
    original_date: string
}


type WeekSelectProps = {
    originalWeekIds: string[]
    setWeekIds: (weekIds: string[]) => void
    leagueId?: string
    update?: boolean
    disabled?: boolean
}

export function WeekSelect({originalWeekIds, setWeekIds, leagueId, update, disabled}: WeekSelectProps) {
    const [weeks, setWeeks] = useState<Week[]>([])
    const {data} = useGetWeeksByLeagueIdQuery(leagueId, {skip: !update})

    const [createWeek] = useCreateWeekMutation()
    const [deleteWeek] = useDeleteWeekMutation()

    useEffect(() => {
        const w = data?.resource
        // console.log("set weeks (retrieved from API)", w)
        setWeeks(w)
    }, [data])

    const weekFormValues = weeks?.map((w: Week) => dayjs(w.date, "YYYY-MM-DD"))

    const onChange = (v: dayjs.Dayjs[]) => {
        // console.log("new value from form:", v)


        const weekDatesYyyyMmDd = v?.map((date) => date.format("YYYY-MM-DD"))
        const weekObjects = weekDatesYyyyMmDd?.map((str): Week => ({
            date: str,
            original_date: str,
        }))

        const newWeeksList: Week[] = []
        for (let newWeek of weekObjects) {
            const existedInOldValues = weeks?.find((oldWeek) => newWeek.date == oldWeek.date)
            if (existedInOldValues) {
                newWeeksList.push(existedInOldValues)
            } else {
                newWeeksList.push(newWeek)
            }
        }

        // console.log("new weeks after form update:", newWeeksList)
        setWeeks(newWeeksList)
    }

    // get all of the week IDs that we had passes
    const getDeletedWeekIds = (): [string[], string[]] => {
        const deletedWeekIds: string[] = []
        const remainingWeekIds: string[] = []
        for (let weekId of originalWeekIds) {
            const existsInNewValues = weeks?.find((week) => week.week_id == weekId)
            if (!existsInNewValues) {
                deletedWeekIds.push(weekId)
            } else {
                remainingWeekIds.push(weekId)
            }
        }

        // console.log("deleted week IDs:", deletedWeekIds)
        // console.log("remaining week IDs:", remainingWeekIds)

        return [deletedWeekIds, remainingWeekIds]
    }

    const getNewWeeks = (): Week[] => {
        const newWeeks: Week[] = []
        for (let i = 0; i < weeks.length; i++) {
            const week = weeks[i]
            if (!week.week_id) {
                newWeeks.push(week)
            }
        }

        // console.log("new weeks:", newWeeks)

        return newWeeks
    }

    const onSave = async () => {

        const [deletedWeekIds, remainingWeekIds] = getDeletedWeekIds()
        for (let weekId of deletedWeekIds) {
            // console.log("delete week: ", weekId)

            const res = await deleteWeek(weekId)
            if (res.error) {
                console.log("error deleting week:", res.error)
            }
        }

        const createdWeekIds: string[] = []
        const newWeeks = getNewWeeks()
        for (let week of newWeeks) {
            const res = await createWeek(week)
            // console.log("create week res:", res)
            if (res.error) {
                console.log("error creating week:", res.error)
            } else {
                createdWeekIds.push(res?.data?.resource?.week_id)
            }
        }


        const newWeekIdsList: string[] = [...remainingWeekIds, ...createdWeekIds]

        // console.log("previous week IDs list:", originalWeekIds)
        // console.log("new week IDs list:", newWeekIdsList)

        setWeekIds(newWeekIdsList)
    }

    return <FormItem name={undefined} label={"Weeks"}>
        <DatePicker multiple minDate={dayjs()} onChange={onChange} value={weekFormValues} disabled={disabled}/>
        {disabled ? null : <Button onClick={onSave}> Save weeks </Button>}
    </FormItem>
}

