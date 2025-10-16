/**
 * Layout principal da aplicação
 * 
 * Autor: Eduardo Jeremias
 * Projeto: INNOVABIZ IAM/TrustGuard
 * Data: 21/08/2025
 */

import React, { ReactNode } from 'react';
import { 
  AppBar, 
  Box, 
  Container, 
  CssBaseline, 
  Drawer, 
  Divider, 
  IconButton, 
  List, 
  ListItem, 
  ListItemButton, 
  ListItemIcon, 
  ListItemText, 
  Toolbar, 
  Typography,
  useTheme
} from '@mui/material';
import { 
  Menu as MenuIcon, 
  ChevronLeft as ChevronLeftIcon, 
  Dashboard as DashboardIcon, 
  Rule as RuleIcon, 
  ViewList as ListIcon, 
  Settings as SettingsIcon, 
  Security as SecurityIcon 
} from '@mui/icons-material';
import { styled } from '@mui/material/styles';
import Link from 'next/link';
import { useRouter } from 'next/router';

// Drawer width
const drawerWidth = 240;

// Styled components
const AppBarStyled = styled(AppBar, {
  shouldForwardProp: (prop) => prop !== 'open',
})<{ open: boolean }>(({ theme, open }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  ...(open && {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  }),
}));

const DrawerStyled = styled(Drawer, {
  shouldForwardProp: (prop) => prop !== 'open',
})(({ theme, open }) => ({
  '& .MuiDrawer-paper': {
    position: 'relative',
    whiteSpace: 'nowrap',
    width: drawerWidth,
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
    boxSizing: 'border-box',
    ...(!open && {
      overflowX: 'hidden',
      transition: theme.transitions.create('width', {
        easing: theme.transitions.easing.sharp,
        duration: theme.transitions.duration.leavingScreen,
      }),
      width: theme.spacing(7),
      [theme.breakpoints.up('sm')]: {
        width: theme.spacing(9),
      },
    }),
  },
}));

/**
 * Estrutura dos itens do menu
 */
interface MenuItem {
  text: string;
  icon: React.ReactNode;
  path: string;
}

/**
 * Propriedades do layout principal
 */
interface MainLayoutProps {
  children: ReactNode;
}

/**
 * Layout principal da aplicação
 */
const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const [open, setOpen] = React.useState(true);
  const router = useRouter();
  const theme = useTheme();
  
  const toggleDrawer = () => {
    setOpen(!open);
  };
  
  // Itens do menu principal
  const menuItems: MenuItem[] = [
    {
      text: 'Dashboard',
      icon: <DashboardIcon />,
      path: '/dashboard',
    },
    {
      text: 'Regras',
      icon: <RuleIcon />,
      path: '/',
    },
    {
      text: 'Conjuntos de Regras',
      icon: <ListIcon />,
      path: '/rule-sets',
    },
    {
      text: 'Segurança',
      icon: <SecurityIcon />,
      path: '/security',
    },
    {
      text: 'Configurações',
      icon: <SettingsIcon />,
      path: '/settings',
    },
  ];
  
  // Verificar se um item está ativo
  const isActive = (path: string) => {
    return router.pathname === path;
  };

  return (
    <Box sx={{ display: 'flex' }}>
      <CssBaseline />
      
      {/* AppBar */}
      <AppBarStyled position="absolute" open={open}>
        <Toolbar
          sx={{
            pr: '24px', // keep right padding when drawer closed
          }}
        >
          <IconButton
            edge="start"
            color="inherit"
            aria-label="open drawer"
            onClick={toggleDrawer}
            sx={{
              marginRight: '36px',
              ...(open && { display: 'none' }),
            }}
          >
            <MenuIcon />
          </IconButton>
          <Typography
            component="h1"
            variant="h6"
            color="inherit"
            noWrap
            sx={{ flexGrow: 1 }}
          >
            INNOVABIZ IAM - Gestão de Regras Dinâmicas
          </Typography>
        </Toolbar>
      </AppBarStyled>
      
      {/* Drawer */}
      <DrawerStyled variant="permanent" open={open}>
        <Toolbar
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            px: [1],
          }}
        >
          <IconButton onClick={toggleDrawer}>
            <ChevronLeftIcon />
          </IconButton>
        </Toolbar>
        <Divider />
        
        {/* Menu Items */}
        <List component="nav">
          {menuItems.map((item) => (
            <Link 
              key={item.text} 
              href={item.path}
              passHref
              style={{ textDecoration: 'none', color: 'inherit' }}
            >
              <ListItem disablePadding>
                <ListItemButton
                  selected={isActive(item.path)}
                  sx={{
                    '&.Mui-selected': {
                      backgroundColor: theme.palette.action.selected,
                      '&:hover': {
                        backgroundColor: theme.palette.action.hover,
                      },
                    },
                  }}
                >
                  <ListItemIcon>
                    {item.icon}
                  </ListItemIcon>
                  <ListItemText primary={item.text} />
                </ListItemButton>
              </ListItem>
            </Link>
          ))}
        </List>
      </DrawerStyled>
      
      {/* Main Content */}
      <Box
        component="main"
        sx={{
          backgroundColor: (theme) =>
            theme.palette.mode === 'light'
              ? theme.palette.grey[100]
              : theme.palette.grey[900],
          flexGrow: 1,
          height: '100vh',
          overflow: 'auto',
        }}
      >
        <Toolbar /> {/* Espaço para compensar AppBar */}
        {children}
      </Box>
    </Box>
  );
};

export default MainLayout;