import {SkillInfoBody, SkillInfoRecord} from "./SkillInfo";
import {CommonFormModal} from "../common/CommonFormModal";
import {
    useCreateSkillInfoMutation,
    useGetSkillInfoOptionsQuery,
} from "../redux/api.js";
import {InputFormItem, NumberInputFormItem, SelectFormItem} from "../common/FormItem";
import * as React from "react";
import {SuggestionInputFormItem} from "../common/SuggestionFormInput";
import {useState} from "react";


type ExperienceFormProps = {
    UserId: string
}

type ExperienceFormOptions = {
    LeagueTypes: string[],
    KnownCaptains: string[],
}

export function ExperienceForm({UserId}: ExperienceFormProps) {

    const [createExperienceRecord] = useCreateSkillInfoMutation()
    const {data} = useGetSkillInfoOptionsQuery()
    const options: ExperienceFormOptions = {
        LeagueTypes: data?.resource?.league_types,
        KnownCaptains: data?.resource?.known_captains,
    }

    const [captain, setCaptain] = useState<string>("")

    const onSubmit = async (formValues: any) => {
        const body: SkillInfoBody = {
            captain: captain,
            league_type: formValues?.league_type,
            level: formValues?.level,
            line: formValues?.line,
            most_recent_year: formValues?.most_recent_year,
            user_id: UserId,
        }

        let func = () => createExperienceRecord(body)
        return await func();
    }

    return <CommonFormModal ObjectType={"experience"} title={"Add info"}
                            IsUpdate={false} InitialState={{}} OnSubmit={onSubmit}>
        <SelectFormItem name={"league_type"} label={"League type"} options={GetLeagueTypes(options)}/>
        <NumberInputFormItem name={"most_recent_year"} label={"Most recent year"}/>
        <SuggestionInputFormItem name={"captain"} label={"Captain"} suggestions={options.KnownCaptains}
                                 value={captain} setValue={setCaptain}/>
        <InputFormItem name={"level"} label={"Level"} placeholder={"e.g. C-4 Seniors"}/>
        <InputFormItem name={"line"} label={"Line"}/>
    </CommonFormModal>
}

function GetLeagueTypes(options: ExperienceFormOptions) {
    return options.LeagueTypes?.map((option => {
        return {label: option, value: option}
    }))
}