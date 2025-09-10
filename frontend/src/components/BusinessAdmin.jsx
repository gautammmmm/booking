import React, { useState, useEffect } from 'react';
import {
  Container,
  AppBar,
  Toolbar,
  Typography,
  Button,
  Box,
  Card,
  CardContent,
  Grid,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  IconButton,
  Alert,
  List,
  ListItem,
  ListItemText,
  CircularProgress,
  ListItemSecondaryAction,
} from '@mui/material';
import { Delete as DeleteIcon, Add as AddIcon } from '@mui/icons-material';
import { servicesAPI, slotsAPI } from '../services/api';

const BusinessAdmin = ({ user, onLogout }) => {
  const [services, setServices] = useState([]); // Initialize as empty array
  const [slots, setSlots] = useState([]); // Initialize as empty array
  const [openServiceDialog, setOpenServiceDialog] = useState(false);
  const [openSlotDialog, setOpenSlotDialog] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  // Service form state
  const [serviceForm, setServiceForm] = useState({
    name: '',
    description: '',
    duration: 30,
  });

  // Slot generation form state
  const [slotForm, setSlotForm] = useState({
    service_id: '',
    start_date: new Date().toISOString().split('T')[0],
    end_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    start_time: '09:00',
    end_time: '17:00',
    interval: 60,
  });

  useEffect(() => {
    loadServices();
    loadSlots();
  }, []);

  const loadServices = async () => {
    try {
      setLoading(true);
      const response = await servicesAPI.list();
      setServices(response.data || []); // Ensure it's always an array
    } catch (error) {
      setError('Failed to load services');
      setServices([]); // Set to empty array on error
    } finally {
      setLoading(false);
    }
  };

  const loadSlots = async () => {
    try {
      setLoading(true);
      const response = await slotsAPI.list();
      setSlots(response.data || []); // Ensure it's always an array
    } catch (error) {
      setError('Failed to load slots');
      setSlots([]); // Set to empty array on error
    } finally {
      setLoading(false);
    }
  };

  const handleCreateService = async () => {
    try {
      await servicesAPI.create(serviceForm);
      setSuccess('Service created successfully!');
      setOpenServiceDialog(false);
      setServiceForm({ name: '', description: '', duration: 30 });
      loadServices();
    } catch (error) {
      setError(error.response?.data?.error || 'Failed to create service');
    }
  };

  const handleDeleteService = async (id) => {
    if (window.confirm('Are you sure you want to delete this service?')) {
      try {
        await servicesAPI.delete(id);
        setSuccess('Service deleted successfully!');
        loadServices();
      } catch (error) {
        setError(error.response?.data?.error || 'Failed to delete service');
      }
    }
  };

  const handleGenerateSlots = async () => {
    try {
      await slotsAPI.generate({
        ...slotForm,
        service_id: parseInt(slotForm.service_id),
        interval: parseInt(slotForm.interval),
      });
      setSuccess('Time slots generated successfully!');
      setOpenSlotDialog(false);
      loadSlots();
    } catch (error) {
      setError(error.response?.data?.error || 'Failed to generate slots');
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Business Admin Dashboard - {user.business?.name || 'My Business'}
          </Typography>
          <Button color="inherit" onClick={onLogout}>
            Logout
          </Button>
        </Toolbar>
      </AppBar>

      <Container sx={{ py: 4 }}>
        {/* Notifications */}
        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
            {error}
          </Alert>
        )}
        {success && (
          <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess('')}>
            {success}
          </Alert>
        )}

        {/* Fix Grid v2 usage - remove item prop and use new breakpoint system */}
        <Grid container spacing={3}>
          {/* Statistics Cards */}
          <Grid size={{ xs: 12, md: 4 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Services
                </Typography>
                <Typography variant="h4">{services.length}</Typography>
                <Button
                  startIcon={<AddIcon />}
                  variant="outlined"
                  sx={{ mt: 2 }}
                  onClick={() => setOpenServiceDialog(true)}
                >
                  Add Service
                </Button>
              </CardContent>
            </Card>
          </Grid>

          <Grid size={{ xs: 12, md: 4 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Time Slots
                </Typography>
                <Typography variant="h4">{slots.length}</Typography>
                <Button
                  startIcon={<AddIcon />}
                  variant="outlined"
                  sx={{ mt: 2 }}
                  onClick={() => setOpenSlotDialog(true)}
                >
                  Generate Slots
                </Button>
              </CardContent>
            </Card>
          </Grid>

          <Grid size={{ xs: 12, md: 4 }}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Available Slots
                </Typography>
                <Typography variant="h4">
                  {slots.filter(slot => slot.is_available).length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Services List */}
        <Box sx={{ mt: 4 }}>
          <Typography variant="h5" gutterBottom>
            Your Services
          </Typography>
          {services.length === 0 ? (
            <Typography color="text.secondary">No services yet. Create your first service!</Typography>
          ) : (
            <List>
              {services.map((service) => (
                <ListItem key={service.id} divider>
                  <ListItemText
                    primary={service.name}
                    secondary={`${service.duration} minutes - ${service.description || 'No description'}`}
                  />
                  <ListItemSecondaryAction>
                    <IconButton
                      edge="end"
                      aria-label="delete"
                      onClick={() => handleDeleteService(service.id)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </ListItemSecondaryAction>
                </ListItem>
              ))}
            </List>
          )}
        </Box>

        {/* Booking Page Link */}
        <Box sx={{ mt: 4 }}>
          <Typography variant="h5" gutterBottom>
            Customer Booking Page
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Share this link with your customers:
          </Typography>
          <Typography
            variant="body2"
            sx={{
              bgcolor: 'grey.100',
              p: 2,
              mt: 1,
              borderRadius: 1,
              fontFamily: 'monospace',
            }}
          >
            {`http://localhost:5173/booking/${user.business_id}`}
          </Typography>
        </Box>

        {/* Service Creation Dialog */}
        <Dialog
          open={openServiceDialog}
          onClose={() => setOpenServiceDialog(false)}
        >
          <DialogTitle>Create New Service</DialogTitle>
          <DialogContent>
            <TextField
              autoFocus
              margin="dense"
              label="Service Name"
              fullWidth
              variant="outlined"
              value={serviceForm.name}
              onChange={(e) =>
                setServiceForm({ ...serviceForm, name: e.target.value })
              }
            />
            <TextField
              margin="dense"
              label="Description"
              fullWidth
              variant="outlined"
              value={serviceForm.description}
              onChange={(e) =>
                setServiceForm({ ...serviceForm, description: e.target.value })
              }
            />
            <TextField
              margin="dense"
              label="Duration (minutes)"
              type="number"
              fullWidth
              variant="outlined"
              value={serviceForm.duration}
              onChange={(e) =>
                setServiceForm({
                  ...serviceForm,
                  duration: parseInt(e.target.value) || 30,
                })
              }
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setOpenServiceDialog(false)}>Cancel</Button>
            <Button onClick={handleCreateService} variant="contained">
              Create Service
            </Button>
          </DialogActions>
        </Dialog>

        {/* Slot Generation Dialog */}
        <Dialog
          open={openSlotDialog}
          onClose={() => setOpenSlotDialog(false)}
          maxWidth="md"
          fullWidth
        >
          <DialogTitle>Generate Time Slots</DialogTitle>
          <DialogContent>
            <Grid container spacing={2} sx={{ mt: 1 }}>
              <Grid size={12}>
                <TextField
                  select
                  label="Service"
                  fullWidth
                  SelectProps={{ native: true }}
                  value={slotForm.service_id}
                  onChange={(e) =>
                    setSlotForm({ ...slotForm, service_id: e.target.value })
                  }
                >
                  <option value="">Select a Service</option>
                  {services.map((service) => (
                    <option key={service.id} value={service.id}>
                      {service.name} ({service.duration} mins)
                    </option>
                  ))}
                </TextField>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <TextField
                  label="Start Date"
                  type="date"
                  fullWidth
                  InputLabelProps={{ shrink: true }}
                  value={slotForm.start_date}
                  onChange={(e) =>
                    setSlotForm({ ...slotForm, start_date: e.target.value })
                  }
                />
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <TextField
                  label="End Date"
                  type="date"
                  fullWidth
                  InputLabelProps={{ shrink: true }}
                  value={slotForm.end_date}
                  onChange={(e) =>
                    setSlotForm({ ...slotForm, end_date: e.target.value })
                  }
                />
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <TextField
                  label="Start Time"
                  type="time"
                  fullWidth
                  InputLabelProps={{ shrink: true }}
                  value={slotForm.start_time}
                  onChange={(e) =>
                    setSlotForm({ ...slotForm, start_time: e.target.value })
                  }
                />
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <TextField
                  label="End Time"
                  type="time"
                  fullWidth
                  InputLabelProps={{ shrink: true }}
                  value={slotForm.end_time}
                  onChange={(e) =>
                    setSlotForm({ ...slotForm, end_time: e.target.value })
                  }
                />
              </Grid>
              <Grid size={12}>
                <TextField
                  label="Interval between slots (minutes)"
                  type="number"
                  fullWidth
                  value={slotForm.interval}
                  onChange={(e) =>
                    setSlotForm({
                      ...slotForm,
                      interval: parseInt(e.target.value) || 60,
                    })
                  }
                />
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setOpenSlotDialog(false)}>Cancel</Button>
            <Button
              onClick={handleGenerateSlots}
              variant="contained"
              disabled={!slotForm.service_id}
            >
              Generate Slots
            </Button>
          </DialogActions>
        </Dialog>
      </Container>
    </>
  );
};

export default BusinessAdmin;