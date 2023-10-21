import Box from '@mui/material/Box';
import Stack from "@mui/material/Stack";

export const Home = () => {
    return (
        <Box>
            <Stack alignItems="center" >
                <Box
                    component="img"
                    sx={{
                        mt: 10
                    }}
                    src="/public/assets/cloud.png"
                />
            </Stack>
        </Box>
    );
}
