import { useState, useEffect } from "react";

import Box from '@mui/material/Box';
import CssBaseline from '@mui/material/CssBaseline';
import Container from '@mui/material/Container';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import TextField from '@mui/material/TextField';

const defaultTheme = createTheme();

const handleAmountFrom = () => console.log("handleAmountFrom");

const App = () => {
    return (
        <ThemeProvider theme={defaultTheme}>
            <Container component="main" maxWidth="sm" sx={{ mt: 2 }}>
                <CssBaseline />
                <Box>
                    <Box sx={{ mb: 3 }}>
                        <TextField fullWidth label="amount" variant="outlined" value="" onChange={handleAmountFrom} />
                    </Box>
                </Box>
            </Container>
        </ThemeProvider>

    );
}

export default App
