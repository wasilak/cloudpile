import Box from '@mui/material/Box';
import CssBaseline from '@mui/material/CssBaseline';
import Container from '@mui/material/Container';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { purple } from '@mui/material/colors';
import ResponsiveAppBar from './menu'

import { Routes, Route, Outlet } from "react-router-dom";

import { Home } from "./pages/home"
import { Search } from "./pages/search"
import { ResourcesList } from "./pages/list"
import { NoMatch } from "./pages/not-found"

const theme = createTheme({
    palette: {
        primary: {
            main: purple[500],
        },
        secondary: {
            main: '#f44336',
        },
    },
});

function Layout() {
    return (
        <Box>
            <ResponsiveAppBar></ResponsiveAppBar>
            <Outlet />
        </Box>
    );
}

const App = () => {
    return (
        <ThemeProvider theme={theme}>
            <Container component="main" maxWidth="xl" sx={{ mt: 2 }}>
                <CssBaseline />

                {/* Routes nest inside one another. Nested route paths build upon parent route paths, and nested route elements render inside parent route elements. See the note about <Outlet> below. */}
                <Routes>
                    <Route path="/" element={<Layout />}>
                        <Route index element={<Home />} />
                        <Route path="search" element={<Search />} />
                        <Route path="list" element={<ResourcesList />} />

                        {/* Using path="*"" means "match anything", so this route acts like a catch-all for URLs that we don't have explicit routes for. */}
                        <Route path="*" element={<NoMatch />} />
                    </Route>
                </Routes>
            </Container>
        </ThemeProvider>
    );
}

export default App
