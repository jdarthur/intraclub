import {useEffect, useState} from "react";
import {
    useCreateWeekMutation,
    useDeleteWeekMutation,
    useGetWeeksByIdsQuery,
} from "../redux/api.js";
import dayjs from "dayjs";
import {FormItem} from "../common/FormItem";
import {DatePicker, notification} from "antd";
import * as React from "react";
import {SubmitResult} from "../common/CommonFormModal";
import {LoadingOutlined} from "@ant-design/icons";

export type Week = {
    week_id?: string
    date: string
    original_date: string
    league_id?: string
}

type WeekSelectProps = {
    originalWeekIds: string[]
    setWeekIds: (w: string[]) => void
    update?: boolean
    disabled?: boolean
}

export type WeekSelectState = {
    deletedWeekIds: string[]
    remainingWeekIds: string[]
    newWeeks: Week[]
}

type CreateWeekResult = {
    createdWeekIds: string[]
    error: { data: { error: any } }
}

export function WeekSelect({...props}: WeekSelectProps) {

    // get all of the week objects from the week IDs list we are passed as props
    const {data, isFetching} = useGetWeeksByIdsQuery({week_ids: props.originalWeekIds}, {skip: !props.update})

    // extract the real Week[] list from the resource key
    const weeks = data?.resource

    const [deleteWeek] = useDeleteWeekMutation()
    const [createWeek] = useCreateWeekMutation()

    // convert all of the weeks into the format we need for the DatePicker
    const weekFormValues = weeks?.map((w: Week) => dayjs(w.date, "YYYY-MM-DD"))

    // notification if we encounter any errors creating / deleting weeks
    const [api, contextHolder] = notification.useNotification();
    const errorNotification = (message: string) => {
        api["error"]({message: 'Error saving week', description: message});
    };

    const onChange = async (v: dayjs.Dayjs[]) => {

        // run through all of the items in the list and calculate the new weeks list
        const [newWeeksList, weeksToCreate] = createWeeksListFromFormValues(v)

        // calculate the deleted and remaining week IDs from the new Week[]
        const [deletedWeekIds, remainingWeekIds] = getDeletedWeekIds(newWeeksList)

        // delete all of the weeks that we removed
        const res = await deleteWeeks(deletedWeekIds)
        if (res.error) {
            errorNotification(res.error.data.error)
            return
        }

        const res2 = await createNewWeeks(weeksToCreate)
        if (res2.error) {
            errorNotification(res.error.data.error)
            return
        }

        const newWeekIds = [...remainingWeekIds, ...res2.createdWeekIds]

        props.setWeekIds(newWeekIds)
    }

    const createWeeksListFromFormValues = (v: dayjs.Dayjs[]) => {

        // format the DatePicker values as YYYY-MM-DD
        const weekDatesYyyyMmDd = v?.map((date) => date.format("YYYY-MM-DD"))

        // create a list of Week[] objects from the YYYY-MM-DD values
        const weekObjects = weekDatesYyyyMmDd?.map((str): Week => ({
            date: str,
            original_date: str,
        })) || []

        const newWeeksList: Week[] = [] // this will be a Week[] containing all of the existing and new weeks
        const weeksToCreate: Week[] = [] // this will be a Week[] containing only the new weeks we need to create via the API

        // run through all of the week objects and determine which ones already existed
        // and which ones will need to be created
        for (let newWeek of weekObjects) {
            const existedInOldValues = weeks?.find((oldWeek: Week) => newWeek.date == oldWeek.date)
            if (existedInOldValues) {
                // if the value existed already it will have a week_id set on it and we won't
                // need to create it, so we can just add it to the "list of all weeks" object
                newWeeksList.push(existedInOldValues)
            } else {
                // otherwise, add the not-yet-created week (without an ID) to the new "list of
                // all weeks" as well as the the "weeks we need to create" list
                newWeeksList.push(newWeek)
                weeksToCreate.push(newWeek)
            }
        }

        return [newWeeksList, weeksToCreate]
    }

    // get all of the week IDs that we have deleted versus the original week IDs passed as props
    const getDeletedWeekIds = (newWeeks: Week[]): [string[], string[]] => {
        const deletedWeekIds: string[] = []
        const remainingWeekIds: string[] = []
        for (let weekId of props.originalWeekIds) {
            // check if this week ID is present in the new list
            const existsInNewValues = newWeeks?.find((week) => week.week_id == weekId)
            if (!existsInNewValues) {
                // if not, add it to the deleted ID list
                deletedWeekIds.push(weekId)
            } else {
                // if so, add it to the remaining IDs list
                remainingWeekIds.push(weekId)
            }
        }

        // return both the deleted week IDs and the non-deleted week IDs
        return [deletedWeekIds, remainingWeekIds]
    }

    // delete all of the weeks in the list, returning any errors that week encounter
    const deleteWeeks = async (weekIds: string[]): Promise<SubmitResult> => {
        for (let weekId of weekIds) {
            const res = await deleteWeek(weekId)
            if (res.error) {
                return res
            }
        }

        return {
            data: null,
            error: null
        }

    }

    // create all of the new weeks from the given list. return an error if we encounter one
    const createNewWeeks = async (w: Week[]): Promise<CreateWeekResult> => {
        const newWeekIds: string[] = []
        for (let week of w) {
            const res = await createWeek(week)
            if (res.error) {
                // delete the weeks that we created before so they are not dangling there
                const res2 = await deleteWeeks(newWeekIds)

                // if we encounter an error while cleaning up, just log it and move on
                if (res2.error) {
                    console.log("error deleting weeks in failed create case: ", res2.error)
                    // return the original error that caused the create failure
                    return res.error
                }

                // return the original error we got when creating the week
                return res.error
            } else {
                // save the new week in a list
                newWeekIds.push(res.data.resource.week_id)
            }
        }

        // return all of the new week IDs so we can set them on the parent component
        return {
            createdWeekIds: newWeekIds,
            error: null
        }
    }

    const loading = <DatePicker><LoadingOutlined/></DatePicker>

    return <FormItem name={undefined} label={"Weeks"}>
        {isFetching ? loading :
            <DatePicker multiple minDate={dayjs()} onChange={onChange}
                        value={weekFormValues} disabled={props.disabled}/>}
    </FormItem>
}
