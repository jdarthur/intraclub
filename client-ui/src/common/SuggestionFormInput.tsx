import {FormItem, NameAndLabel} from "./FormItem";
import {AutoComplete} from "antd";
import React from "react";
import {useState} from "react";

type SuggestionInputFormItemProps = NameAndLabel & {
    suggestions: string[]
    value: string
    setValue: (value: string) => void
}

type Option = {
    label: string
    value: string
}

export function SuggestionInputFormItem({
                                            name,
                                            label,
                                            disabled,
                                            placeholder,
                                            suggestions,
                                            value,
                                            setValue,
                                        }: SuggestionInputFormItemProps) {

    const [filteredOptions, setFilteredOptions] = useState<Option[]>([]);

    const onChange = (value: string) => {
        setValue(value);
        setFilteredOptions(getFilteredOptions());
    }

    const getFilteredOptions = (): Option[] => {
        const options: Option[] = [];
        for (const option of suggestions) {
            if (option.toLowerCase().includes(value.toLowerCase())) {
                options.push({label: option, value: option});
            }
        }
        return options;
    }
    return <FormItem name={name} label={label} disabled={disabled} placeholder={placeholder}>
        <AutoComplete placeholder={placeholder} options={filteredOptions} value={value} onChange={onChange}/>
    </FormItem>
}
