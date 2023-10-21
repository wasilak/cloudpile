import Box from '@mui/material/Box';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import InputLabel from '@mui/material/InputLabel';

export const FilterSelect = ({ items, selected, onChange, label }: any) => {

    const uniqueId = (Date.now() * Math.random()).toString();

    return (
        <Box>
            <FormControl fullWidth sx={{ mb: 1 }}>
                <InputLabel id={uniqueId}>{label}</InputLabel>
                <Select
                    labelId={uniqueId}
                    value={selected}
                    label={label}
                    onChange={onChange}
                >
                    {items &&
                        items.map(function (item: string) {
                            return (
                                <MenuItem key={item} value={item}>{item}</MenuItem>
                            );
                        })
                    }
                </Select>
            </FormControl>
        </Box>
    );
}
