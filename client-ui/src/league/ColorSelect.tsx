import * as React from "react";
import {Button, ColorPicker, Input, Popover, Space, Tooltip} from "antd";
import {DeleteOutlined, PlusSquareOutlined} from "@ant-design/icons";
import {FormItem} from "../common/FormItem";
import {useState} from "react";
import {TeamColor} from "../team/TeamColor";
import {Color} from "antd/es/color-picker";
import {c} from "vite/dist/node/types.d-aGj9QkWt";
import {Simulate} from "react-dom/test-utils";
import select = Simulate.select;


export type TeamColor = {
    name: string
    hex: string
}

type ColorSelectProps = {
    colors: TeamColor[]
    setColors: (c: TeamColor[]) => void
    disabled?: boolean
}

const addColorStyle = {
    width: 200,
    backgroundColor: "#f5f5f5",
    padding: "1em",
    borderRadius: 5,
    display: "flex",
    flexDirection: "column" as "column"
}

export function ColorSelect({colors, setColors, disabled}: ColorSelectProps) {

    const [showAddColor, setShowAddColor] = useState<boolean>(false)
    const [color, setColor] = useState<TeamColor>({hex: undefined, name: ""})
    const [selected, setSelected] = useState<number>(-1)

    const editColor = (color: TeamColor) => {

        let selectedIndex = -1;

        for (let i = 0; i < colors?.length; i++) {
            const c = colors[i]
            if (c.hex == color.hex) {
                selectedIndex = i
                break
            }
        }

        setSelected(selectedIndex)

        setColor(color)
        setShowAddColor(true)
    }

    const onSave = () => {
        const c: TeamColor = {
            name: color.name,
            hex: color.hex
        }

        const newColors = [...colors]
        if (selected === -2) {
            newColors.push(c)
        } else {
            newColors[selected] = c
        }

        setColors(newColors)

        setColor({hex: undefined, name: ""})
        setShowAddColor(false)
    }

    const deleteColor = (index: number) => {
        const c = [...colors]
        c.splice(index, 1)
        setColors(c)
        setShowAddColor(false)
    }

    const selectedColors = colors?.map((c, i) => {
        const display = <div onClick={() => editColor(c)}>
            <ColorDisplay name={c.name} hex={c.hex}
                          key={c.hex} disabled={disabled} editable/>
        </div>

        const deleteFunc = () => deleteColor(i)

        if (showAddColor && selected == i && !disabled) {
            return <TeamColorSelector color={color} save={onSave} setColor={setColor} update
                                      cancel={() => setShowAddColor(false)}
                                      button={display} deleteColor={deleteFunc}/>
        }

        return display
    })

    const addNew = () => {
        setColor({hex: undefined, name: undefined})
        setSelected(-2)
        setShowAddColor(true)
    }

    const addColorButton = <Button onClick={addNew} style={{marginTop: "0.5em"}}>
        <PlusSquareOutlined/>
        Add
    </Button>

    const addColorBox = showAddColor && selected == -2 ?
        <TeamColorSelector color={color} save={onSave} setColor={setColor}
                           cancel={() => setShowAddColor(false)}
                           button={addColorButton} disabled={disabled}/> : null


    return <FormItem name={""} label={"Team Colors"}>
        <Space style={{display: "flex", flexWrap: "wrap", marginBottom: "1em"}}>
            {selectedColors}
        </Space>
        {addColorBox}
        {(showAddColor || disabled) ? null : addColorButton}
    </FormItem>
}

type teamColorSelectorProps = {
    color: TeamColor
    setColor: (c: TeamColor) => void
    save: () => void
    cancel: () => void
    button: React.ReactNode
    deleteColor?: () => void // required only when update == true
    update?: boolean
    disabled?: boolean
}

function TeamColorSelector({
                               color,
                               setColor,
                               save,
                               cancel,
                               deleteColor,
                               button,
                               update,
                               disabled
                           }: teamColorSelectorProps) {

    const onColorChange = (v: any) => {
        const c: TeamColor = {...color}
        c.hex = v.toHex()
        setColor(c)
    }

    const onTeamNameChange = (e: any) => {
        const c: TeamColor = {...color}
        c.name = e.target.value
        setColor(c)
    }

    const body = <div style={addColorStyle}>
        <Space style={{display: "flex", flexDirection: "row", marginBottom: "0.5em"}}>
            <ColorPicker onChange={onColorChange} value={color.hex}/>
            <Input placeholder={"Team name"} onChange={onTeamNameChange} value={color.name}/>
        </Space>

        <Space style={{alignSelf: "flex-end"}}>
            <Button onClick={cancel}> Cancel </Button>
            <Button onClick={save} type={"primary"}> Save </Button>
        </Space>
    </div>

    const editTitle = <div style={{display: "flex", justifyContent: "space-between"}}>
        Edit team color
        <DeleteOutlined onClick={deleteColor}/>
    </div>

    const title = update ? editTitle : "Choose team color"

    return <Popover open title={title} content={body}>
        {button}
    </Popover>
}

type colorDisplayProps = TeamColor & {
    disabled?: boolean
    editable?: boolean
}

export function ColorDisplay({name, hex, disabled, editable}: colorDisplayProps) {
    const backgroundColor = disabled ? "#f0f0f0" : "white"
    const borderColor = disabled ? "rgba(191, 191, 191, 0.8)" : "#d9d9d9"

    const style = {
        padding: 3,
        borderRadius: 5,
        border: `1px solid ${borderColor}`,
        backgroundColor: backgroundColor
    }
    if (editable && !disabled) {
        style["cursor"] = "pointer"
    }


    return <Tooltip title={name}>
        <div
            style={style}>
            < div style={{borderRadius: 5, backgroundColor: `#${hex}`, height: 25, width: 25}}/>
        </div>

    </Tooltip>
}